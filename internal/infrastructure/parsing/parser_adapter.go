package parsing

import (
	"gitlab.com/amoconst/germinator/internal/application"
)

// parsingParser implements the application.Parser interface.
type parsingParser struct{}

// NewParser creates a new Parser instance that implements the application.Parser interface.
func NewParser() application.Parser {
	return &parsingParser{}
}

// LoadDocument loads and parses a document from the given path.
func (p *parsingParser) LoadDocument(path string, platform string) (interface{}, error) {
	return LoadDocument(path, platform)
}

// Compile-time interface satisfaction check.
var _ application.Parser = (*parsingParser)(nil)
