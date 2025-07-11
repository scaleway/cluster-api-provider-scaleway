// Code generated by MockGen. DO NOT EDIT.
// Source: ../domain.go
//
// Generated by this command:
//
//	mockgen -destination domain_mock.go -package mock_client -source ../domain.go -typed
//

// Package mock_client is a generated GoMock package.
package mock_client

import (
	context "context"
	reflect "reflect"

	domain "github.com/scaleway/scaleway-sdk-go/api/domain/v2beta1"
	scw "github.com/scaleway/scaleway-sdk-go/scw"
	gomock "go.uber.org/mock/gomock"
)

// MockDomainAPI is a mock of DomainAPI interface.
type MockDomainAPI struct {
	ctrl     *gomock.Controller
	recorder *MockDomainAPIMockRecorder
	isgomock struct{}
}

// MockDomainAPIMockRecorder is the mock recorder for MockDomainAPI.
type MockDomainAPIMockRecorder struct {
	mock *MockDomainAPI
}

// NewMockDomainAPI creates a new mock instance.
func NewMockDomainAPI(ctrl *gomock.Controller) *MockDomainAPI {
	mock := &MockDomainAPI{ctrl: ctrl}
	mock.recorder = &MockDomainAPIMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockDomainAPI) EXPECT() *MockDomainAPIMockRecorder {
	return m.recorder
}

// ListDNSZoneRecords mocks base method.
func (m *MockDomainAPI) ListDNSZoneRecords(req *domain.ListDNSZoneRecordsRequest, opts ...scw.RequestOption) (*domain.ListDNSZoneRecordsResponse, error) {
	m.ctrl.T.Helper()
	varargs := []any{req}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "ListDNSZoneRecords", varargs...)
	ret0, _ := ret[0].(*domain.ListDNSZoneRecordsResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListDNSZoneRecords indicates an expected call of ListDNSZoneRecords.
func (mr *MockDomainAPIMockRecorder) ListDNSZoneRecords(req any, opts ...any) *MockDomainAPIListDNSZoneRecordsCall {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{req}, opts...)
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListDNSZoneRecords", reflect.TypeOf((*MockDomainAPI)(nil).ListDNSZoneRecords), varargs...)
	return &MockDomainAPIListDNSZoneRecordsCall{Call: call}
}

// MockDomainAPIListDNSZoneRecordsCall wrap *gomock.Call
type MockDomainAPIListDNSZoneRecordsCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockDomainAPIListDNSZoneRecordsCall) Return(arg0 *domain.ListDNSZoneRecordsResponse, arg1 error) *MockDomainAPIListDNSZoneRecordsCall {
	c.Call = c.Call.Return(arg0, arg1)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockDomainAPIListDNSZoneRecordsCall) Do(f func(*domain.ListDNSZoneRecordsRequest, ...scw.RequestOption) (*domain.ListDNSZoneRecordsResponse, error)) *MockDomainAPIListDNSZoneRecordsCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockDomainAPIListDNSZoneRecordsCall) DoAndReturn(f func(*domain.ListDNSZoneRecordsRequest, ...scw.RequestOption) (*domain.ListDNSZoneRecordsResponse, error)) *MockDomainAPIListDNSZoneRecordsCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// UpdateDNSZoneRecords mocks base method.
func (m *MockDomainAPI) UpdateDNSZoneRecords(req *domain.UpdateDNSZoneRecordsRequest, opts ...scw.RequestOption) (*domain.UpdateDNSZoneRecordsResponse, error) {
	m.ctrl.T.Helper()
	varargs := []any{req}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "UpdateDNSZoneRecords", varargs...)
	ret0, _ := ret[0].(*domain.UpdateDNSZoneRecordsResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// UpdateDNSZoneRecords indicates an expected call of UpdateDNSZoneRecords.
func (mr *MockDomainAPIMockRecorder) UpdateDNSZoneRecords(req any, opts ...any) *MockDomainAPIUpdateDNSZoneRecordsCall {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{req}, opts...)
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateDNSZoneRecords", reflect.TypeOf((*MockDomainAPI)(nil).UpdateDNSZoneRecords), varargs...)
	return &MockDomainAPIUpdateDNSZoneRecordsCall{Call: call}
}

// MockDomainAPIUpdateDNSZoneRecordsCall wrap *gomock.Call
type MockDomainAPIUpdateDNSZoneRecordsCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockDomainAPIUpdateDNSZoneRecordsCall) Return(arg0 *domain.UpdateDNSZoneRecordsResponse, arg1 error) *MockDomainAPIUpdateDNSZoneRecordsCall {
	c.Call = c.Call.Return(arg0, arg1)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockDomainAPIUpdateDNSZoneRecordsCall) Do(f func(*domain.UpdateDNSZoneRecordsRequest, ...scw.RequestOption) (*domain.UpdateDNSZoneRecordsResponse, error)) *MockDomainAPIUpdateDNSZoneRecordsCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockDomainAPIUpdateDNSZoneRecordsCall) DoAndReturn(f func(*domain.UpdateDNSZoneRecordsRequest, ...scw.RequestOption) (*domain.UpdateDNSZoneRecordsResponse, error)) *MockDomainAPIUpdateDNSZoneRecordsCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// MockDomain is a mock of Domain interface.
type MockDomain struct {
	ctrl     *gomock.Controller
	recorder *MockDomainMockRecorder
	isgomock struct{}
}

// MockDomainMockRecorder is the mock recorder for MockDomain.
type MockDomainMockRecorder struct {
	mock *MockDomain
}

// NewMockDomain creates a new mock instance.
func NewMockDomain(ctrl *gomock.Controller) *MockDomain {
	mock := &MockDomain{ctrl: ctrl}
	mock.recorder = &MockDomainMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockDomain) EXPECT() *MockDomainMockRecorder {
	return m.recorder
}

// DeleteDNSZoneRecords mocks base method.
func (m *MockDomain) DeleteDNSZoneRecords(ctx context.Context, zone, name string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteDNSZoneRecords", ctx, zone, name)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteDNSZoneRecords indicates an expected call of DeleteDNSZoneRecords.
func (mr *MockDomainMockRecorder) DeleteDNSZoneRecords(ctx, zone, name any) *MockDomainDeleteDNSZoneRecordsCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteDNSZoneRecords", reflect.TypeOf((*MockDomain)(nil).DeleteDNSZoneRecords), ctx, zone, name)
	return &MockDomainDeleteDNSZoneRecordsCall{Call: call}
}

// MockDomainDeleteDNSZoneRecordsCall wrap *gomock.Call
type MockDomainDeleteDNSZoneRecordsCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockDomainDeleteDNSZoneRecordsCall) Return(arg0 error) *MockDomainDeleteDNSZoneRecordsCall {
	c.Call = c.Call.Return(arg0)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockDomainDeleteDNSZoneRecordsCall) Do(f func(context.Context, string, string) error) *MockDomainDeleteDNSZoneRecordsCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockDomainDeleteDNSZoneRecordsCall) DoAndReturn(f func(context.Context, string, string) error) *MockDomainDeleteDNSZoneRecordsCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// ListDNSZoneRecords mocks base method.
func (m *MockDomain) ListDNSZoneRecords(ctx context.Context, zone, name string) ([]*domain.Record, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListDNSZoneRecords", ctx, zone, name)
	ret0, _ := ret[0].([]*domain.Record)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListDNSZoneRecords indicates an expected call of ListDNSZoneRecords.
func (mr *MockDomainMockRecorder) ListDNSZoneRecords(ctx, zone, name any) *MockDomainListDNSZoneRecordsCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListDNSZoneRecords", reflect.TypeOf((*MockDomain)(nil).ListDNSZoneRecords), ctx, zone, name)
	return &MockDomainListDNSZoneRecordsCall{Call: call}
}

// MockDomainListDNSZoneRecordsCall wrap *gomock.Call
type MockDomainListDNSZoneRecordsCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockDomainListDNSZoneRecordsCall) Return(arg0 []*domain.Record, arg1 error) *MockDomainListDNSZoneRecordsCall {
	c.Call = c.Call.Return(arg0, arg1)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockDomainListDNSZoneRecordsCall) Do(f func(context.Context, string, string) ([]*domain.Record, error)) *MockDomainListDNSZoneRecordsCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockDomainListDNSZoneRecordsCall) DoAndReturn(f func(context.Context, string, string) ([]*domain.Record, error)) *MockDomainListDNSZoneRecordsCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// SetDNSZoneRecords mocks base method.
func (m *MockDomain) SetDNSZoneRecords(ctx context.Context, zone, name string, ips []string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetDNSZoneRecords", ctx, zone, name, ips)
	ret0, _ := ret[0].(error)
	return ret0
}

// SetDNSZoneRecords indicates an expected call of SetDNSZoneRecords.
func (mr *MockDomainMockRecorder) SetDNSZoneRecords(ctx, zone, name, ips any) *MockDomainSetDNSZoneRecordsCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetDNSZoneRecords", reflect.TypeOf((*MockDomain)(nil).SetDNSZoneRecords), ctx, zone, name, ips)
	return &MockDomainSetDNSZoneRecordsCall{Call: call}
}

// MockDomainSetDNSZoneRecordsCall wrap *gomock.Call
type MockDomainSetDNSZoneRecordsCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockDomainSetDNSZoneRecordsCall) Return(arg0 error) *MockDomainSetDNSZoneRecordsCall {
	c.Call = c.Call.Return(arg0)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockDomainSetDNSZoneRecordsCall) Do(f func(context.Context, string, string, []string) error) *MockDomainSetDNSZoneRecordsCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockDomainSetDNSZoneRecordsCall) DoAndReturn(f func(context.Context, string, string, []string) error) *MockDomainSetDNSZoneRecordsCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}
