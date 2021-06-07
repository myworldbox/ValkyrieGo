// Code generated by mockery v2.8.0. DO NOT EDIT.

package mocks

import (
	model "github.com/sentrionic/valkyrie/model"
	mock "github.com/stretchr/testify/mock"
)

// MessageRepository is an autogenerated mock type for the MessageRepository type
type MessageRepository struct {
	mock.Mock
}

// CreateMessage provides a mock function with given fields: message
func (_m *MessageRepository) CreateMessage(message *model.Message) error {
	ret := _m.Called(message)

	var r0 error
	if rf, ok := ret.Get(0).(func(*model.Message) error); ok {
		r0 = rf(message)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DeleteMessage provides a mock function with given fields: message
func (_m *MessageRepository) DeleteMessage(message *model.Message) error {
	ret := _m.Called(message)

	var r0 error
	if rf, ok := ret.Get(0).(func(*model.Message) error); ok {
		r0 = rf(message)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetById provides a mock function with given fields: messageId
func (_m *MessageRepository) GetById(messageId string) (*model.Message, error) {
	ret := _m.Called(messageId)

	var r0 *model.Message
	if rf, ok := ret.Get(0).(func(string) *model.Message); ok {
		r0 = rf(messageId)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Message)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(messageId)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetMessages provides a mock function with given fields: userId, channel, cursor
func (_m *MessageRepository) GetMessages(userId string, channel *model.Channel, cursor string) (*[]model.MessageResponse, error) {
	ret := _m.Called(userId, channel, cursor)

	var r0 *[]model.MessageResponse
	if rf, ok := ret.Get(0).(func(string, *model.Channel, string) *[]model.MessageResponse); ok {
		r0 = rf(userId, channel, cursor)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*[]model.MessageResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, *model.Channel, string) error); ok {
		r1 = rf(userId, channel, cursor)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// UpdateMessage provides a mock function with given fields: message
func (_m *MessageRepository) UpdateMessage(message *model.Message) error {
	ret := _m.Called(message)

	var r0 error
	if rf, ok := ret.Get(0).(func(*model.Message) error); ok {
		r0 = rf(message)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}