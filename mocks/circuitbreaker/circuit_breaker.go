// Code generated by mockery v2.41.0. DO NOT EDIT.

package circuitbreaker

import (
	circuitbreaker "github.com/kanthorlabs/common/circuitbreaker"
	mock "github.com/stretchr/testify/mock"
)

// CircuitBreaker is an autogenerated mock type for the CircuitBreaker type
type CircuitBreaker struct {
	mock.Mock
}

type CircuitBreaker_Expecter struct {
	mock *mock.Mock
}

func (_m *CircuitBreaker) EXPECT() *CircuitBreaker_Expecter {
	return &CircuitBreaker_Expecter{mock: &_m.Mock}
}

// Do provides a mock function with given fields: cmd, onHandle, onError
func (_m *CircuitBreaker) Do(cmd string, onHandle circuitbreaker.Handler, onError circuitbreaker.ErrorHandler) (interface{}, error) {
	ret := _m.Called(cmd, onHandle, onError)

	if len(ret) == 0 {
		panic("no return value specified for Do")
	}

	var r0 interface{}
	var r1 error
	if rf, ok := ret.Get(0).(func(string, circuitbreaker.Handler, circuitbreaker.ErrorHandler) (interface{}, error)); ok {
		return rf(cmd, onHandle, onError)
	}
	if rf, ok := ret.Get(0).(func(string, circuitbreaker.Handler, circuitbreaker.ErrorHandler) interface{}); ok {
		r0 = rf(cmd, onHandle, onError)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(interface{})
		}
	}

	if rf, ok := ret.Get(1).(func(string, circuitbreaker.Handler, circuitbreaker.ErrorHandler) error); ok {
		r1 = rf(cmd, onHandle, onError)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// CircuitBreaker_Do_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Do'
type CircuitBreaker_Do_Call struct {
	*mock.Call
}

// Do is a helper method to define mock.On call
//   - cmd string
//   - onHandle circuitbreaker.Handler
//   - onError circuitbreaker.ErrorHandler
func (_e *CircuitBreaker_Expecter) Do(cmd interface{}, onHandle interface{}, onError interface{}) *CircuitBreaker_Do_Call {
	return &CircuitBreaker_Do_Call{Call: _e.mock.On("Do", cmd, onHandle, onError)}
}

func (_c *CircuitBreaker_Do_Call) Run(run func(cmd string, onHandle circuitbreaker.Handler, onError circuitbreaker.ErrorHandler)) *CircuitBreaker_Do_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string), args[1].(circuitbreaker.Handler), args[2].(circuitbreaker.ErrorHandler))
	})
	return _c
}

func (_c *CircuitBreaker_Do_Call) Return(_a0 interface{}, _a1 error) *CircuitBreaker_Do_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *CircuitBreaker_Do_Call) RunAndReturn(run func(string, circuitbreaker.Handler, circuitbreaker.ErrorHandler) (interface{}, error)) *CircuitBreaker_Do_Call {
	_c.Call.Return(run)
	return _c
}

// NewCircuitBreaker creates a new instance of CircuitBreaker. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewCircuitBreaker(t interface {
	mock.TestingT
	Cleanup(func())
}) *CircuitBreaker {
	mock := &CircuitBreaker{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
