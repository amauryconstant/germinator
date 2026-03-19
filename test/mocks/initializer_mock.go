package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
	"gitlab.com/amoconst/germinator/internal/application"
	"gitlab.com/amoconst/germinator/internal/domain"
)

// MockInitializer is a mock implementation of application.Initializer.
type MockInitializer struct {
	mock.Mock
}

// Initialize provides a mock function with given fields: ctx, req.
func (_m *MockInitializer) Initialize(ctx context.Context, req *application.InitializeRequest) ([]domain.InitializeResult, error) {
	ret := _m.Called(ctx, req)

	var r0 []domain.InitializeResult
	if rf, ok := ret.Get(0).(func(context.Context, *application.InitializeRequest) []domain.InitializeResult); ok {
		r0 = rf(ctx, req)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]domain.InitializeResult)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *application.InitializeRequest) error); ok {
		r1 = rf(ctx, req)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
