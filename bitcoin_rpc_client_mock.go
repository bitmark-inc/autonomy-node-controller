// Code generated by MockGen. DO NOT EDIT.
// Source: bitcoin_rpc_client.go

// Package main is a generated GoMock package.
package main

import (
	json "encoding/json"
	reflect "reflect"

	btcjson "github.com/btcsuite/btcd/btcjson"
	rpcclient "github.com/btcsuite/btcd/rpcclient"
	gomock "github.com/golang/mock/gomock"
)

// MockrpcClient is a mock of rpcClient interface.
type MockrpcClient struct {
	ctrl     *gomock.Controller
	recorder *MockrpcClientMockRecorder
}

// MockrpcClientMockRecorder is the mock recorder for MockrpcClient.
type MockrpcClientMockRecorder struct {
	mock *MockrpcClient
}

// NewMockrpcClient creates a new mock instance.
func NewMockrpcClient(ctrl *gomock.Controller) *MockrpcClient {
	mock := &MockrpcClient{ctrl: ctrl}
	mock.recorder = &MockrpcClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockrpcClient) EXPECT() *MockrpcClientMockRecorder {
	return m.recorder
}

// GetBlockChainInfo mocks base method.
func (m *MockrpcClient) GetBlockChainInfo() (*btcjson.GetBlockChainInfoResult, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetBlockChainInfo")
	ret0, _ := ret[0].(*btcjson.GetBlockChainInfoResult)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetBlockChainInfo indicates an expected call of GetBlockChainInfo.
func (mr *MockrpcClientMockRecorder) GetBlockChainInfo() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetBlockChainInfo", reflect.TypeOf((*MockrpcClient)(nil).GetBlockChainInfo))
}

// GetDescriptorInfo mocks base method.
func (m *MockrpcClient) GetDescriptorInfo(descriptor string) (*btcjson.GetDescriptorInfoResult, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetDescriptorInfo", descriptor)
	ret0, _ := ret[0].(*btcjson.GetDescriptorInfoResult)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetDescriptorInfo indicates an expected call of GetDescriptorInfo.
func (mr *MockrpcClientMockRecorder) GetDescriptorInfo(descriptor interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetDescriptorInfo", reflect.TypeOf((*MockrpcClient)(nil).GetDescriptorInfo), descriptor)
}

// GetWalletInfo mocks base method.
func (m *MockrpcClient) GetWalletInfo() (*btcjson.GetWalletInfoResult, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetWalletInfo")
	ret0, _ := ret[0].(*btcjson.GetWalletInfoResult)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetWalletInfo indicates an expected call of GetWalletInfo.
func (mr *MockrpcClientMockRecorder) GetWalletInfo() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetWalletInfo", reflect.TypeOf((*MockrpcClient)(nil).GetWalletInfo))
}

// RawRequest mocks base method.
func (m *MockrpcClient) RawRequest(method string, params []json.RawMessage) (json.RawMessage, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RawRequest", method, params)
	ret0, _ := ret[0].(json.RawMessage)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// RawRequest indicates an expected call of RawRequest.
func (mr *MockrpcClientMockRecorder) RawRequest(method, params interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RawRequest", reflect.TypeOf((*MockrpcClient)(nil).RawRequest), method, params)
}

// Shutdown mocks base method.
func (m *MockrpcClient) Shutdown() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Shutdown")
}

// Shutdown indicates an expected call of Shutdown.
func (mr *MockrpcClientMockRecorder) Shutdown() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Shutdown", reflect.TypeOf((*MockrpcClient)(nil).Shutdown))
}

// WalletProcessPsbt mocks base method.
func (m *MockrpcClient) WalletProcessPsbt(psbt string, sign *bool, sighashType rpcclient.SigHashType, bip32Derivs *bool) (*btcjson.WalletProcessPsbtResult, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "WalletProcessPsbt", psbt, sign, sighashType, bip32Derivs)
	ret0, _ := ret[0].(*btcjson.WalletProcessPsbtResult)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// WalletProcessPsbt indicates an expected call of WalletProcessPsbt.
func (mr *MockrpcClientMockRecorder) WalletProcessPsbt(psbt, sign, sighashType, bip32Derivs interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WalletProcessPsbt", reflect.TypeOf((*MockrpcClient)(nil).WalletProcessPsbt), psbt, sign, sighashType, bip32Derivs)
}
