// Code generated by mockery v2.50.0. DO NOT EDIT.

package service

import (
	mock "github.com/stretchr/testify/mock"

	uuid "github.com/google/uuid"
)

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

// GetAvatar provides a mock function with given fields: nickname
func (_m *MockRandomizer) GetAvatar(nickname string) string {
	ret := _m.Called(nickname)

	if len(ret) == 0 {
		panic("no return value specified for GetAvatar")
	}

	var r0 string
	if rf, ok := ret.Get(0).(func(string) string); ok {
		r0 = rf(nickname)
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// MockRandomizer_GetAvatar_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetAvatar'
type MockRandomizer_GetAvatar_Call struct {
	*mock.Call
}

// GetAvatar is a helper method to define mock.On call
//   - nickname string
func (_e *MockRandomizer_Expecter) GetAvatar(nickname interface{}) *MockRandomizer_GetAvatar_Call {
	return &MockRandomizer_GetAvatar_Call{Call: _e.mock.On("GetAvatar", nickname)}
}

func (_c *MockRandomizer_GetAvatar_Call) Run(run func(nickname string)) *MockRandomizer_GetAvatar_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *MockRandomizer_GetAvatar_Call) Return(_a0 string) *MockRandomizer_GetAvatar_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockRandomizer_GetAvatar_Call) RunAndReturn(run func(string) string) *MockRandomizer_GetAvatar_Call {
	_c.Call.Return(run)
	return _c
}

// GetFibberIndex provides a mock function with given fields: playersLen
func (_m *MockRandomizer) GetFibberIndex(playersLen int) int {
	ret := _m.Called(playersLen)

	if len(ret) == 0 {
		panic("no return value specified for GetFibberIndex")
	}

	var r0 int
	if rf, ok := ret.Get(0).(func(int) int); ok {
		r0 = rf(playersLen)
	} else {
		r0 = ret.Get(0).(int)
	}

	return r0
}

// MockRandomizer_GetFibberIndex_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetFibberIndex'
type MockRandomizer_GetFibberIndex_Call struct {
	*mock.Call
}

// GetFibberIndex is a helper method to define mock.On call
//   - playersLen int
func (_e *MockRandomizer_Expecter) GetFibberIndex(playersLen interface{}) *MockRandomizer_GetFibberIndex_Call {
	return &MockRandomizer_GetFibberIndex_Call{Call: _e.mock.On("GetFibberIndex", playersLen)}
}

func (_c *MockRandomizer_GetFibberIndex_Call) Run(run func(playersLen int)) *MockRandomizer_GetFibberIndex_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(int))
	})
	return _c
}

func (_c *MockRandomizer_GetFibberIndex_Call) Return(_a0 int) *MockRandomizer_GetFibberIndex_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockRandomizer_GetFibberIndex_Call) RunAndReturn(run func(int) int) *MockRandomizer_GetFibberIndex_Call {
	_c.Call.Return(run)
	return _c
}

// GetID provides a mock function with no fields
func (_m *MockRandomizer) GetID() uuid.UUID {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for GetID")
	}

	var r0 uuid.UUID
	if rf, ok := ret.Get(0).(func() uuid.UUID); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(uuid.UUID)
		}
	}

	return r0
}

// MockRandomizer_GetID_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetID'
type MockRandomizer_GetID_Call struct {
	*mock.Call
}

// GetID is a helper method to define mock.On call
func (_e *MockRandomizer_Expecter) GetID() *MockRandomizer_GetID_Call {
	return &MockRandomizer_GetID_Call{Call: _e.mock.On("GetID")}
}

func (_c *MockRandomizer_GetID_Call) Run(run func()) *MockRandomizer_GetID_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockRandomizer_GetID_Call) Return(_a0 uuid.UUID) *MockRandomizer_GetID_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockRandomizer_GetID_Call) RunAndReturn(run func() uuid.UUID) *MockRandomizer_GetID_Call {
	_c.Call.Return(run)
	return _c
}

// GetNickname provides a mock function with no fields
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

// GetRoomCode provides a mock function with no fields
func (_m *MockRandomizer) GetRoomCode() string {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for GetRoomCode")
	}

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// MockRandomizer_GetRoomCode_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetRoomCode'
type MockRandomizer_GetRoomCode_Call struct {
	*mock.Call
}

// GetRoomCode is a helper method to define mock.On call
func (_e *MockRandomizer_Expecter) GetRoomCode() *MockRandomizer_GetRoomCode_Call {
	return &MockRandomizer_GetRoomCode_Call{Call: _e.mock.On("GetRoomCode")}
}

func (_c *MockRandomizer_GetRoomCode_Call) Run(run func()) *MockRandomizer_GetRoomCode_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockRandomizer_GetRoomCode_Call) Return(_a0 string) *MockRandomizer_GetRoomCode_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockRandomizer_GetRoomCode_Call) RunAndReturn(run func() string) *MockRandomizer_GetRoomCode_Call {
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
