package output

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"
	"text/tabwriter"

	"gitlab.com/amoconst/germinator/internal/iostreams"
)

// Exporter writes data to an IOStreams destination in a specific format.
type Exporter interface {
	Write(io *iostreams.IOStreams, data any) error
}

// JSONExporter writes data as pretty-printed JSON (2-space indent,
// trailing newline).
type JSONExporter struct{}

// NewJSONExporter constructs a JSONExporter.
func NewJSONExporter() *JSONExporter { return &JSONExporter{} }

// Write marshals data as JSON and writes it to io.Out.
func (e *JSONExporter) Write(io *iostreams.IOStreams, data any) error {
	buf, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("json marshal: %w", err)
	}
	buf = append(buf, '\n')
	if _, err := io.Out.Write(buf); err != nil {
		return fmt.Errorf("json write: %w", err)
	}
	return nil
}

// TableExporter writes a slice of structs as a tab-aligned text table.
// Struct fields are read in declaration order; the `tab:"HEADER"` tag
// controls the column header (defaulting to the field name).
type TableExporter struct{}

// NewTableExporter constructs a TableExporter.
func NewTableExporter() *TableExporter { return &TableExporter{} }

// Write renders data as a tab-aligned table to io.Out.
func (e *TableExporter) Write(io *iostreams.IOStreams, data any) error {
	rows, err := buildRows(data)
	if err != nil {
		return err
	}
	if len(rows) == 0 {
		return nil
	}
	tw := tabwriter.NewWriter(io.Out, 0, 0, 2, ' ', 0)
	for _, row := range rows {
		if _, err := fmt.Fprintln(tw, row); err != nil {
			return fmt.Errorf("table write: %w", err)
		}
	}
	if err := tw.Flush(); err != nil {
		return fmt.Errorf("table flush: %w", err)
	}
	return nil
}

func buildRows(data any) ([]string, error) {
	v := reflect.ValueOf(data)
	for v.Kind() == reflect.Pointer {
		v = v.Elem()
	}
	if v.Kind() != reflect.Slice {
		return nil, fmt.Errorf("table exporter: expected slice, got %T", data)
	}
	n := v.Len()
	if n == 0 {
		return nil, nil
	}
	first := indirectValue(v.Index(0))
	headers, fieldIndexes, err := headersFor(first)
	if err != nil {
		return nil, err
	}
	rows := make([]string, 0, n+1)
	rows = append(rows, strings.Join(headers, "\t"))
	for i := 0; i < n; i++ {
		elem := indirectValue(v.Index(i))
		cells := cellsFor(elem, fieldIndexes)
		rows = append(rows, strings.Join(cells, "\t"))
	}
	return rows, nil
}

func indirectValue(v reflect.Value) reflect.Value {
	for v.Kind() == reflect.Pointer {
		v = v.Elem()
	}
	return v
}

func headersFor(v reflect.Value) ([]string, []int, error) {
	if v.Kind() != reflect.Struct {
		return nil, nil, fmt.Errorf("table exporter: slice element must be a struct, got %s", v.Kind())
	}
	t := v.Type()
	headers := make([]string, 0, t.NumField())
	idxs := make([]int, 0, t.NumField())
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if !f.IsExported() {
			continue
		}
		header := f.Tag.Get("tab")
		if header == "" {
			header = f.Name
		} else if header == "-" {
			continue
		}
		headers = append(headers, header)
		idxs = append(idxs, i)
	}
	return headers, idxs, nil
}

func cellsFor(v reflect.Value, idxs []int) []string {
	cells := make([]string, 0, len(idxs))
	for _, i := range idxs {
		f := v.Field(i)
		cells = append(cells, formatCell(f))
	}
	return cells
}

func formatCell(v reflect.Value) string {
	if !v.IsValid() {
		return ""
	}
	if v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return ""
		}
		v = v.Elem()
	}
	switch v.Kind() {
	case reflect.String:
		return v.String()
	case reflect.Bool:
		if v.Bool() {
			return "true"
		}
		return "false"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(v.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(v.Uint(), 10)
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(v.Float(), 'g', -1, 64)
	default:
		b, err := json.Marshal(v.Interface())
		if err != nil {
			return fmt.Sprintf("%v", v.Interface())
		}
		var buf bytes.Buffer
		_ = json.Compact(&buf, b)
		return buf.String()
	}
}

var _ = io.EOF
