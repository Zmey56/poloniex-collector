// Code generated by MockGen. DO NOT EDIT.
// Source: internal/domain/repository/interfaces.go

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	models "github.com/Zmey56/poloniex-collector/internal/domain/models"
	gomock "github.com/golang/mock/gomock"
)

// MockTradeRepository is a mock of TradeRepository interface.
type MockTradeRepository struct {
	ctrl     *gomock.Controller
	recorder *MockTradeRepositoryMockRecorder
}

// MockTradeRepositoryMockRecorder is the mock recorder for MockTradeRepository.
type MockTradeRepositoryMockRecorder struct {
	mock *MockTradeRepository
}

// NewMockTradeRepository creates a new mock instance.
func NewMockTradeRepository(ctrl *gomock.Controller) *MockTradeRepository {
	mock := &MockTradeRepository{ctrl: ctrl}
	mock.recorder = &MockTradeRepositoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockTradeRepository) EXPECT() *MockTradeRepositoryMockRecorder {
	return m.recorder
}

// SaveTrade mocks base method.
func (m *MockTradeRepository) SaveTrade(ctx context.Context, trade models.RecentTrade) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SaveTrade", ctx, trade)
	ret0, _ := ret[0].(error)
	return ret0
}

// SaveTrade indicates an expected call of SaveTrade.
func (mr *MockTradeRepositoryMockRecorder) SaveTrade(ctx, trade interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SaveTrade", reflect.TypeOf((*MockTradeRepository)(nil).SaveTrade), ctx, trade)
}

// SaveTrades mocks base method.
func (m *MockTradeRepository) SaveTrades(ctx context.Context, trades []models.RecentTrade) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SaveTrades", ctx, trades)
	ret0, _ := ret[0].(error)
	return ret0
}

// SaveTrades indicates an expected call of SaveTrades.
func (mr *MockTradeRepositoryMockRecorder) SaveTrades(ctx, trades interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SaveTrades", reflect.TypeOf((*MockTradeRepository)(nil).SaveTrades), ctx, trades)
}

// MockKlineRepository is a mock of KlineRepository interface.
type MockKlineRepository struct {
	ctrl     *gomock.Controller
	recorder *MockKlineRepositoryMockRecorder
}

// MockKlineRepositoryMockRecorder is the mock recorder for MockKlineRepository.
type MockKlineRepositoryMockRecorder struct {
	mock *MockKlineRepository
}

// NewMockKlineRepository creates a new mock instance.
func NewMockKlineRepository(ctrl *gomock.Controller) *MockKlineRepository {
	mock := &MockKlineRepository{ctrl: ctrl}
	mock.recorder = &MockKlineRepositoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockKlineRepository) EXPECT() *MockKlineRepositoryMockRecorder {
	return m.recorder
}

// GetKlineByInterval mocks base method.
func (m *MockKlineRepository) GetKlineByInterval(ctx context.Context, pair, timeframe string, beginTime int64) (*models.Kline, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetKlineByInterval", ctx, pair, timeframe, beginTime)
	ret0, _ := ret[0].(*models.Kline)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetKlineByInterval indicates an expected call of GetKlineByInterval.
func (mr *MockKlineRepositoryMockRecorder) GetKlineByInterval(ctx, pair, timeframe, beginTime interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetKlineByInterval", reflect.TypeOf((*MockKlineRepository)(nil).GetKlineByInterval), ctx, pair, timeframe, beginTime)
}

// GetKlinesByTimeRange mocks base method.
func (m *MockKlineRepository) GetKlinesByTimeRange(ctx context.Context, pair, timeframe string, startTime, endTime int64) ([]models.Kline, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetKlinesByTimeRange", ctx, pair, timeframe, startTime, endTime)
	ret0, _ := ret[0].([]models.Kline)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetKlinesByTimeRange indicates an expected call of GetKlinesByTimeRange.
func (mr *MockKlineRepositoryMockRecorder) GetKlinesByTimeRange(ctx, pair, timeframe, startTime, endTime interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetKlinesByTimeRange", reflect.TypeOf((*MockKlineRepository)(nil).GetKlinesByTimeRange), ctx, pair, timeframe, startTime, endTime)
}

// GetLastKline mocks base method.
func (m *MockKlineRepository) GetLastKline(ctx context.Context, pair, timeframe string) (*models.Kline, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetLastKline", ctx, pair, timeframe)
	ret0, _ := ret[0].(*models.Kline)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetLastKline indicates an expected call of GetLastKline.
func (mr *MockKlineRepositoryMockRecorder) GetLastKline(ctx, pair, timeframe interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetLastKline", reflect.TypeOf((*MockKlineRepository)(nil).GetLastKline), ctx, pair, timeframe)
}

// SaveKline mocks base method.
func (m *MockKlineRepository) SaveKline(ctx context.Context, kline models.Kline) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SaveKline", ctx, kline)
	ret0, _ := ret[0].(error)
	return ret0
}

// SaveKline indicates an expected call of SaveKline.
func (mr *MockKlineRepositoryMockRecorder) SaveKline(ctx, kline interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SaveKline", reflect.TypeOf((*MockKlineRepository)(nil).SaveKline), ctx, kline)
}

// MockExchangeClient is a mock of ExchangeClient interface.
type MockExchangeClient struct {
	ctrl     *gomock.Controller
	recorder *MockExchangeClientMockRecorder
}

// MockExchangeClientMockRecorder is the mock recorder for MockExchangeClient.
type MockExchangeClientMockRecorder struct {
	mock *MockExchangeClient
}

// NewMockExchangeClient creates a new mock instance.
func NewMockExchangeClient(ctrl *gomock.Controller) *MockExchangeClient {
	mock := &MockExchangeClient{ctrl: ctrl}
	mock.recorder = &MockExchangeClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockExchangeClient) EXPECT() *MockExchangeClientMockRecorder {
	return m.recorder
}

// GetHistoricalKlines mocks base method.
func (m *MockExchangeClient) GetHistoricalKlines(ctx context.Context, pair, timeframe string, startTime, endTime int64) ([]models.Kline, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetHistoricalKlines", ctx, pair, timeframe, startTime, endTime)
	ret0, _ := ret[0].([]models.Kline)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetHistoricalKlines indicates an expected call of GetHistoricalKlines.
func (mr *MockExchangeClientMockRecorder) GetHistoricalKlines(ctx, pair, timeframe, startTime, endTime interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetHistoricalKlines", reflect.TypeOf((*MockExchangeClient)(nil).GetHistoricalKlines), ctx, pair, timeframe, startTime, endTime)
}

// SubscribeToTrades mocks base method.
func (m *MockExchangeClient) SubscribeToTrades(ctx context.Context, pairs []string) (<-chan models.RecentTrade, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SubscribeToTrades", ctx, pairs)
	ret0, _ := ret[0].(<-chan models.RecentTrade)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SubscribeToTrades indicates an expected call of SubscribeToTrades.
func (mr *MockExchangeClientMockRecorder) SubscribeToTrades(ctx, pairs interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SubscribeToTrades", reflect.TypeOf((*MockExchangeClient)(nil).SubscribeToTrades), ctx, pairs)
}
