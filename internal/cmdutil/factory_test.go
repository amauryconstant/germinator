package cmdutil

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/amoconst/germinator/internal/application"
	"gitlab.com/amoconst/germinator/internal/config"
	"gitlab.com/amoconst/germinator/internal/core"
	"gitlab.com/amoconst/germinator/internal/iostreams"
	"gitlab.com/amoconst/germinator/internal/library"
)

func newFactory() *Factory {
	return NewFactory(iostreams.Test(), "test", "germinator")
}

func TestFactoryEagerFields(t *testing.T) {
	t.Parallel()

	io := iostreams.Test()
	f := NewFactory(io, "1.2.3", "germinator")
	assert.Equal(t, io, f.IOStreams)
	assert.Equal(t, "1.2.3", f.AppVersion)
	assert.Equal(t, "germinator", f.Executable)
	assert.NotNil(t, f.RootContext)
	f.Close()
}

func TestFactoryLazyFieldCaching(t *testing.T) {
	t.Parallel()

	var counter int
	f := newFactory()
	f.Library = OnceValuesFunc(func() (*library.Library, error) {
		counter++
		return &library.Library{}, nil
	})

	v1, err := f.Library()
	require.NoError(t, err)
	v2, err := f.Library()
	require.NoError(t, err)
	assert.Same(t, v1, v2)
	assert.Equal(t, 1, counter, "library function should be called exactly once")
}

func TestFactoryTwoCallersShareCachedValue(t *testing.T) {
	t.Parallel()

	var counter int
	f := newFactory()
	f.Library = OnceValuesFunc(func() (*library.Library, error) {
		counter++
		return &library.Library{Version: "shared"}, nil
	})

	v1, err := f.Library()
	require.NoError(t, err)
	v2, err := f.Library()
	require.NoError(t, err)
	assert.Same(t, v1, v2)
	assert.Equal(t, "shared", v1.Version)
	assert.Equal(t, 1, counter)
}

func TestFactoryConcurrentFirstCallOnce(t *testing.T) {
	t.Parallel()

	var counter int32
	f := newFactory()
	f.Library = OnceValuesFunc(func() (*library.Library, error) {
		atomic.AddInt32(&counter, 1)
		return &library.Library{Version: "concurrent"}, nil
	})

	var wg sync.WaitGroup
	const n = 50
	results := make([]*library.Library, n)
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			v, err := f.Library()
			require.NoError(t, err)
			results[idx] = v
		}(i)
	}
	wg.Wait()

	assert.Equal(t, int32(1), atomic.LoadInt32(&counter), "function should be called exactly once under concurrency")
	for i := 1; i < n; i++ {
		assert.Same(t, results[0], results[i], "all callers should receive the same instance")
	}
}

func TestFactoryErrorCached(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("disk full")
	var counter int
	f := newFactory()
	f.Library = OnceValuesFunc(func() (*library.Library, error) {
		counter++
		return nil, wantErr
	})

	_, err1 := f.Library()
	_, err2 := f.Library()
	assert.Equal(t, wantErr, err1)
	assert.Equal(t, wantErr, err2)
	assert.Equal(t, 1, counter, "function should be called exactly once even when it errors")
}

type stubTransformer struct{ name string }

func (s *stubTransformer) Transform(_ context.Context, _ *application.TransformRequest) (*core.TransformResult, error) {
	return &core.TransformResult{}, nil
}

func TestFactoryCrossDependencyCaching(t *testing.T) {
	t.Parallel()

	var configCount, libCount, transformerCount, initCount int

	f := newFactory()
	f.Config = OnceValuesFunc(func() (*config.Config, error) {
		configCount++
		return &config.Config{}, nil
	})
	f.Library = OnceValuesFunc(func() (*library.Library, error) {
		libCount++
		if _, err := f.Config(); err != nil {
			return nil, err
		}
		return &library.Library{Version: "v1"}, nil
	})
	f.Transformer = OnceValuesFunc(func() (application.Transformer, error) {
		transformerCount++
		if _, err := f.Config(); err != nil {
			return nil, err
		}
		return &stubTransformer{name: "stub"}, nil
	})
	f.Initializer = OnceValuesFunc(func() (application.Initializer, error) {
		initCount++
		if _, err := f.Library(); err != nil {
			return nil, err
		}
		return nil, nil
	})

	_, _ = f.Initializer()
	_, _ = f.Library()
	_, _ = f.Transformer()

	assert.Equal(t, 1, configCount, "Config function should be called exactly once")
	assert.Equal(t, 1, libCount)
	assert.Equal(t, 1, transformerCount)
	assert.Equal(t, 1, initCount)
}

func TestFactoryInstancesHaveIndependentCaches(t *testing.T) {
	t.Parallel()

	var counter int
	mk := func() *Factory {
		f := newFactory()
		f.Library = OnceValuesFunc(func() (*library.Library, error) {
			counter++
			return &library.Library{Version: "fresh"}, nil
		})
		return f
	}

	f1 := mk()
	f2 := mk()

	v1, err := f1.Library()
	require.NoError(t, err)
	v2, err := f2.Library()
	require.NoError(t, err)

	assert.Equal(t, 2, counter, "each Factory instance must invoke its function independently")
	assert.NotSame(t, v1, v2, "each Factory must yield its own instance")
}

func TestFactoryAllLazyFieldsCounted(t *testing.T) {
	t.Parallel()

	var validatorCount, canonicalizerCount int
	f := newFactory()
	f.Validator = OnceValuesFunc(func() (application.Validator, error) {
		validatorCount++
		return nil, nil
	})
	f.Canonicalizer = OnceValuesFunc(func() (application.Canonicalizer, error) {
		canonicalizerCount++
		return nil, nil
	})
	f.Initializer = OnceValuesFunc(func() (application.Initializer, error) { return nil, nil })
	f.Library = OnceValuesFunc(func() (*library.Library, error) { return &library.Library{}, nil })
	f.Transformer = OnceValuesFunc(func() (application.Transformer, error) { return nil, nil })
	f.Config = OnceValuesFunc(func() (*config.Config, error) { return &config.Config{}, nil })

	_, _ = f.Initializer()
	_, _ = f.Library()
	_, _ = f.Transformer()
	_, _ = f.Validator()
	_, _ = f.Canonicalizer()
	_, _ = f.Config()

	assert.Equal(t, 1, validatorCount, "Validator function should be called exactly once")
	assert.Equal(t, 1, canonicalizerCount, "Canonicalizer function should be called exactly once")
}
