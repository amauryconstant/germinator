package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
	"gitlab.com/amoconst/germinator/internal/application"
	"gitlab.com/amoconst/germinator/internal/core"
)

// MockTransformer is a mock implementation of application.Transformer.
type MockTransformer struct {
	mock.Mock
}

// Transform provides a mock function with given fields: ctx, req.
func (_m *MockTransformer) Transform(ctx context.Context, req *application.TransformRequest) (*core.TransformResult, error) {
	ret := _m.Called(ctx, req)

	var r0 *core.TransformResult
	if rf, ok := ret.Get(0).(func(context.Context, *application.TransformRequest) *core.TransformResult); ok {
		r0 = rf(ctx, req)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*core.TransformResult)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *application.TransformRequest) error); ok {
		r1 = rf(ctx, req)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
