package mockclient

import (
	"bytes"
	"context"
	"io"
	"net/http"
)

type MockResp struct {
	Data []byte
	Code int
}

type MockRespList []MockResp

type contextKey string

const (
	contextMockClient = contextKey("mockclient")
)

// MockClient is the mock client
type MockClient struct {
	doFunc   func(req *http.Request) (*http.Response, error)
	mockData MockRespList
	index    int
}

func defaultDoFunc(req *http.Request) (*http.Response, error) {
	ctxValue := req.Context().Value(contextMockClient)
	if ctxValue == nil {
		panic("no MockClient context")
	}
	mc, ok := ctxValue.(*MockClient)
	if !ok {
		panic("returned value is not a MockClient!")
	}
	if mc == nil {
		panic("no data")
	}
	if mc.index == len(mc.mockData) {
		panic("ran out of data")
	}
	resp := &http.Response{
		StatusCode: mc.mockData[mc.index].Code,
		Body:       io.NopCloser(bytes.NewReader([]byte(mc.mockData[mc.index].Data))),
		Header:     make(http.Header),
	}
	mc.index++
	return resp, nil
}

// Do is the mock client's `Do` func
func (m *MockClient) Do(req *http.Request) (*http.Response, error) {
	return m.doFunc(req)
}

func (m *MockClient) SetDoFunc(df func(req *http.Request) (*http.Response, error)) {
	m.doFunc = df
}

func (m *MockClient) Reset() {
	m.index = 0
}

func (m *MockClient) SetData(data MockRespList) {
	m.index = 0
	m.mockData = data
}

func (m *MockClient) GetData() MockRespList {
	return m.mockData
}

func (m *MockClient) GetIndex() int {
	return m.index
}

func (m *MockClient) LastData() []byte {
	if m.index == 0 {
		return m.mockData[0].Data
	}
	return m.mockData[m.index-1].Data
}

func (m *MockClient) Empty() bool {
	return m.index == len(m.mockData)
}

func NewMockClient() (*MockClient, context.Context) {
	mc := &MockClient{
		doFunc:   defaultDoFunc,
		mockData: nil,
		index:    0,
	}
	return mc, context.WithValue(context.TODO(), contextMockClient, mc)
}
