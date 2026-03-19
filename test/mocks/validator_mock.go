package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
	"gitlab.com/amoconst/germinator/internal/application"
	"gitlab.com/amoconst/germinator/internal/domain"
)

// MockValidator is a mock implementation of application.Validator.
type MockValidator struct {
	mock.Mock
}

// Validate provides a mock function with given fields: ctx, req.
func (_m *MockValidator) Validate(ctx context.Context, req *application.ValidateRequest) (*domain.ValidateResult, error) {
	ret := _m.Called(ctx, req)

	var r0 *domain.ValidateResult
	if rf, ok := ret.Get(0).(func(context.Context, *application.ValidateRequest) *domain.ValidateResult); ok {
		r0 = rf(ctx, req)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*domain.ValidateResult)
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
