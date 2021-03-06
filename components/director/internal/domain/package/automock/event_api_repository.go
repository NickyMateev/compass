// Code generated by mockery v1.0.0. DO NOT EDIT.

package automock

import context "context"
import mock "github.com/stretchr/testify/mock"
import model "github.com/kyma-incubator/compass/components/director/internal/model"

// EventAPIRepository is an autogenerated mock type for the EventAPIRepository type
type EventAPIRepository struct {
	mock.Mock
}

// Create provides a mock function with given fields: ctx, items
func (_m *EventAPIRepository) Create(ctx context.Context, items *model.EventDefinition) error {
	ret := _m.Called(ctx, items)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *model.EventDefinition) error); ok {
		r0 = rf(ctx, items)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
