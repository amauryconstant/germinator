package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
	"gitlab.com/amoconst/germinator/internal/application"
	"gitlab.com/amoconst/germinator/internal/domain"
)

// MockCanonicalizer is a mock implementation of application.Canonicalizer.
type MockCanonicalizer struct {
	mock.Mock
}

// Canonicalize provides a mock function with given fields: ctx, req.
func (_m *MockCanonicalizer) Canonicalize(ctx context.Context, req *application.CanonicalizeRequest) (*domain.CanonicalizeResult, error) {
	ret := _m.Called(ctx, req)

	var r0 *domain.CanonicalizeResult
	if rf, ok := ret.Get(0).(func(context.Context, *application.CanonicalizeRequest) *domain.CanonicalizeResult); ok {
		r0 = rf(ctx, req)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*domain.CanonicalizeResult)
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
