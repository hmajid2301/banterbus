// Code generated by mockery v2.46.0. DO NOT EDIT.

package service

import mock "github.com/stretchr/testify/mock"

// MockRandomizer is an autogenerated mock type for the Randomizer type
type MockRandomizer struct {
	mock.Mock
}

type MockRandomizer_Expecter struct {
	mock *mock.Mock
}

func (_m *MockRandomizer) EXPECT() *MockRandomizer_Expecter {
	return &MockRandomizer_Expecter{mock: &_m.Mock}
}

// GetAvatar provides a mock function with given fields:
func (_m *MockRandomizer) GetAvatar() []byte {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for GetAvatar")
	}

	var r0 []byte
	if rf, ok := ret.Get(0).(func() []byte); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]byte)
		}
	}

	return r0
}

// MockRandomizer_GetAvatar_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetAvatar'
type MockRandomizer_GetAvatar_Call struct {
	*mock.Call
}

// GetAvatar is a helper method to define mock.On call
func (_e *MockRandomizer_Expecter) GetAvatar() *MockRandomizer_GetAvatar_Call {
	return &MockRandomizer_GetAvatar_Call{Call: _e.mock.On("GetAvatar")}
}

func (_c *MockRandomizer_GetAvatar_Call) Run(run func()) *MockRandomizer_GetAvatar_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockRandomizer_GetAvatar_Call) Return(_a0 []byte) *MockRandomizer_GetAvatar_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockRandomizer_GetAvatar_Call) RunAndReturn(run func() []byte) *MockRandomizer_GetAvatar_Call {
	_c.Call.Return(run)
	return _c
}

// GetNickname provides a mock function with given fields:
func (_m *MockRandomizer) GetNickname() string {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for GetNickname")
	}

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// MockRandomizer_GetNickname_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetNickname'
type MockRandomizer_GetNickname_Call struct {
	*mock.Call
}

// GetNickname is a helper method to define mock.On call
func (_e *MockRandomizer_Expecter) GetNickname() *MockRandomizer_GetNickname_Call {
	return &MockRandomizer_GetNickname_Call{Call: _e.mock.On("GetNickname")}
}

func (_c *MockRandomizer_GetNickname_Call) Run(run func()) *MockRandomizer_GetNickname_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockRandomizer_GetNickname_Call) Return(_a0 string) *MockRandomizer_GetNickname_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockRandomizer_GetNickname_Call) RunAndReturn(run func() string) *MockRandomizer_GetNickname_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockRandomizer creates a new instance of MockRandomizer. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockRandomizer(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockRandomizer {
	mock := &MockRandomizer{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
