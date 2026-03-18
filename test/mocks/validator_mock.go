package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
	"gitlab.com/amoconst/germinator/internal/application"
)

// MockValidator is a mock implementation of application.Validator.
type MockValidator struct {
	mock.Mock
}

// Validate provides a mock function with given fields: ctx, req.
func (_m *MockValidator) Validate(ctx context.Context, req *application.ValidateRequest) (*application.ValidateResult, error) {
	ret := _m.Called(ctx, req)

	var r0 *application.ValidateResult
	if rf, ok := ret.Get(0).(func(context.Context, *application.ValidateRequest) *application.ValidateResult); ok {
		r0 = rf(ctx, req)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*application.ValidateResult)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *application.ValidateRequest) error); ok {
		r1 = rf(ctx, req)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
