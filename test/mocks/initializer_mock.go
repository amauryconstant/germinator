package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
	"gitlab.com/amoconst/germinator/internal/application"
)

// MockInitializer is a mock implementation of application.Initializer.
type MockInitializer struct {
	mock.Mock
}

// Initialize provides a mock function with given fields: ctx, req.
func (_m *MockInitializer) Initialize(ctx context.Context, req *application.InitializeRequest) ([]application.InitializeResult, error) {
	ret := _m.Called(ctx, req)

	var r0 []application.InitializeResult
	if rf, ok := ret.Get(0).(func(context.Context, *application.InitializeRequest) []application.InitializeResult); ok {
		r0 = rf(ctx, req)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]application.InitializeResult)
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
