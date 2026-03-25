package mocks

import (
	"github.com/stretchr/testify/mock"
	"gitlab.com/amoconst/germinator/internal/application"
)

// MockSerializer is a mock implementation of application.Serializer.
type MockSerializer struct {
	mock.Mock
}

// RenderDocument provides a mock function with given fields: doc, platform.
func (_m *MockSerializer) RenderDocument(doc interface{}, platform string) (string, error) {
	ret := _m.Called(doc, platform)

	var r0 string
	if rf, ok := ret.Get(0).(func(interface{}, string) string); ok {
		r0 = rf(doc, platform)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(interface{}, string) error); ok {
		r1 = rf(doc, platform)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Compile-time interface satisfaction check.
var _ application.Serializer = (*MockSerializer)(nil)
