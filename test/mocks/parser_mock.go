package mocks

import (
	"github.com/stretchr/testify/mock"
	"gitlab.com/amoconst/germinator/internal/application"
)

// MockParser is a mock implementation of application.Parser.
type MockParser struct {
	mock.Mock
}

// LoadDocument provides a mock function with given fields: path, platform.
func (_m *MockParser) LoadDocument(path string, platform string) (interface{}, error) {
	ret := _m.Called(path, platform)

	var r0 interface{}
	if rf, ok := ret.Get(0).(func(string, string) interface{}); ok {
		r0 = rf(path, platform)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string) error); ok {
		r1 = rf(path, platform)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Compile-time interface satisfaction check.
var _ application.Parser = (*MockParser)(nil)
