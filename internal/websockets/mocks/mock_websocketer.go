// Code generated by mockery v2.52.1. DO NOT EDIT.

package websockets

import (
	context "context"

	redis "github.com/redis/go-redis/v9"
	mock "github.com/stretchr/testify/mock"

	uuid "github.com/google/uuid"
)

// MockWebsocketer is an autogenerated mock type for the Websocketer type
type MockWebsocketer struct {
	mock.Mock
}

type MockWebsocketer_Expecter struct {
	mock *mock.Mock
}

func (_m *MockWebsocketer) EXPECT() *MockWebsocketer_Expecter {
	return &MockWebsocketer_Expecter{mock: &_m.Mock}
}

// Close provides a mock function with given fields: id
func (_m *MockWebsocketer) Close(id uuid.UUID) error {
	ret := _m.Called(id)

	if len(ret) == 0 {
		panic("no return value specified for Close")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(uuid.UUID) error); ok {
		r0 = rf(id)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockWebsocketer_Close_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Close'
type MockWebsocketer_Close_Call struct {
	*mock.Call
}

// Close is a helper method to define mock.On call
//   - id uuid.UUID
func (_e *MockWebsocketer_Expecter) Close(id interface{}) *MockWebsocketer_Close_Call {
	return &MockWebsocketer_Close_Call{Call: _e.mock.On("Close", id)}
}

func (_c *MockWebsocketer_Close_Call) Run(run func(id uuid.UUID)) *MockWebsocketer_Close_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(uuid.UUID))
	})
	return _c
}

func (_c *MockWebsocketer_Close_Call) Return(_a0 error) *MockWebsocketer_Close_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockWebsocketer_Close_Call) RunAndReturn(run func(uuid.UUID) error) *MockWebsocketer_Close_Call {
	_c.Call.Return(run)
	return _c
}

// Publish provides a mock function with given fields: ctx, id, msg
func (_m *MockWebsocketer) Publish(ctx context.Context, id uuid.UUID, msg []byte) error {
	ret := _m.Called(ctx, id, msg)

	if len(ret) == 0 {
		panic("no return value specified for Publish")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, []byte) error); ok {
		r0 = rf(ctx, id, msg)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockWebsocketer_Publish_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Publish'
type MockWebsocketer_Publish_Call struct {
	*mock.Call
}

// Publish is a helper method to define mock.On call
//   - ctx context.Context
//   - id uuid.UUID
//   - msg []byte
func (_e *MockWebsocketer_Expecter) Publish(ctx interface{}, id interface{}, msg interface{}) *MockWebsocketer_Publish_Call {
	return &MockWebsocketer_Publish_Call{Call: _e.mock.On("Publish", ctx, id, msg)}
}

func (_c *MockWebsocketer_Publish_Call) Run(run func(ctx context.Context, id uuid.UUID, msg []byte)) *MockWebsocketer_Publish_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(uuid.UUID), args[2].([]byte))
	})
	return _c
}

func (_c *MockWebsocketer_Publish_Call) Return(_a0 error) *MockWebsocketer_Publish_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockWebsocketer_Publish_Call) RunAndReturn(run func(context.Context, uuid.UUID, []byte) error) *MockWebsocketer_Publish_Call {
	_c.Call.Return(run)
	return _c
}

// Subscribe provides a mock function with given fields: ctx, id
func (_m *MockWebsocketer) Subscribe(ctx context.Context, id uuid.UUID) <-chan *redis.Message {
	ret := _m.Called(ctx, id)

	if len(ret) == 0 {
		panic("no return value specified for Subscribe")
	}

	var r0 <-chan *redis.Message
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) <-chan *redis.Message); ok {
		r0 = rf(ctx, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(<-chan *redis.Message)
		}
	}

	return r0
}

// MockWebsocketer_Subscribe_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Subscribe'
type MockWebsocketer_Subscribe_Call struct {
	*mock.Call
}

// Subscribe is a helper method to define mock.On call
//   - ctx context.Context
//   - id uuid.UUID
func (_e *MockWebsocketer_Expecter) Subscribe(ctx interface{}, id interface{}) *MockWebsocketer_Subscribe_Call {
	return &MockWebsocketer_Subscribe_Call{Call: _e.mock.On("Subscribe", ctx, id)}
}

func (_c *MockWebsocketer_Subscribe_Call) Run(run func(ctx context.Context, id uuid.UUID)) *MockWebsocketer_Subscribe_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(uuid.UUID))
	})
	return _c
}

func (_c *MockWebsocketer_Subscribe_Call) Return(_a0 <-chan *redis.Message) *MockWebsocketer_Subscribe_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockWebsocketer_Subscribe_Call) RunAndReturn(run func(context.Context, uuid.UUID) <-chan *redis.Message) *MockWebsocketer_Subscribe_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockWebsocketer creates a new instance of MockWebsocketer. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockWebsocketer(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockWebsocketer {
	mock := &MockWebsocketer{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
