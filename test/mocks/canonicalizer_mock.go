package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
	"gitlab.com/amoconst/germinator/internal/application"
)

// MockCanonicalizer is a mock implementation of application.Canonicalizer.
type MockCanonicalizer struct {
	mock.Mock
}

// Canonicalize provides a mock function with given fields: ctx, req.
func (_m *MockCanonicalizer) Canonicalize(ctx context.Context, req *application.CanonicalizeRequest) (*application.CanonicalizeResult, error) {
	ret := _m.Called(ctx, req)

	var r0 *application.CanonicalizeResult
	if rf, ok := ret.Get(0).(func(context.Context, *application.CanonicalizeRequest) *application.CanonicalizeResult); ok {
		r0 = rf(ctx, req)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*application.CanonicalizeResult)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *application.CanonicalizeRequest) error); ok {
		r1 = rf(ctx, req)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
