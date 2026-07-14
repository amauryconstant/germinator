package cmdutil

import (
	"context"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/amoconst/germinator/internal/iostreams"
)

func TestFactoryEagerFields(t *testing.T) {
	t.Parallel()

	io := iostreams.Test()
	f := NewFactory(context.Background(), io, "1.2.3", "germinator")
	assert.Equal(t, io, f.IOStreams)
	assert.Equal(t, "1.2.3", f.AppVersion)
	assert.Equal(t, "germinator", f.Executable)
	assert.NotNil(t, f.RootContext)
	f.Close()
}

// TestFactory_OnlyConfigAndLibraryAreLazyFields enforces the
// cli-cli-factory/spec.md "No additional lazy service fields" scenario
// via reflection: the Factory struct MUST expose exactly Config and
// Library as func()-typed EXPORTED fields. Adding Transformer, Validator,
// Canonicalizer, or Initializer would fail this test.
//
// Note: rootCancel is an unexported context.CancelFunc for the Factory's
// internal Close() lifecycle. Unexported fields are ignored — only
// exported func fields are subject to the contract (callers cannot
// depend on unexported fields anyway).
func TestFactory_OnlyConfigAndLibraryAreLazyFields(t *testing.T) {
	t.Parallel()

	factoryType := reflect.TypeOf(Factory{})
	var lazyFuncs []string
	for i := 0; i < factoryType.NumField(); i++ {
		f := factoryType.Field(i)
		if !f.IsExported() {
			continue
		}
		if f.Type.Kind() != reflect.Func {
			continue
		}
		lazyFuncs = append(lazyFuncs, f.Name)
	}

	assert.ElementsMatch(t, []string{"Config", "Library"}, lazyFuncs,
		"Factory MUST expose exactly Config and Library as lazy func() fields")
}

// TestFactory_LazyFieldTypes verifies the func signatures of the
// allowed lazy fields so a future type change to either field is
// caught at test time.
func TestFactory_LazyFieldTypes(t *testing.T) {
	t.Parallel()

	factoryType := reflect.TypeOf(Factory{})

	configField, ok := factoryType.FieldByName("Config")
	require.True(t, ok, "Factory MUST have a Config field")
	assert.Equal(t, "func() (*config.Config, error)", configField.Type.String(),
		"Factory.Config MUST be func() (*config.Config, error)")

	libraryField, ok := factoryType.FieldByName("Library")
	require.True(t, ok, "Factory MUST have a Library field")
	assert.Equal(t, "func() (*library.Library, error)", libraryField.Type.String(),
		"Factory.Library MUST be func() (*library.Library, error)")
}
