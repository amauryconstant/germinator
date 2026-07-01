package cmdutil

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/amoconst/germinator/internal/iostreams"
	"gitlab.com/amoconst/germinator/internal/library"
)

func newFactory() *Factory {
	return NewFactory(context.Background(), iostreams.Test(), "test", "germinator")
}

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
