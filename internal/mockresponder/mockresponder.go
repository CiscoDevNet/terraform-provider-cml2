package mockclient

import (
	"bytes"
	"context"
	"io"
	"log"
	"net/http"
	"regexp"
	"sync"
)

type MockResp struct {
	Data   []byte
	Code   int
	URL    string
	served bool
}

type MockRespList []MockResp

type contextKey string

const (
	contextMockClient = contextKey("mockclient")
)

// MockResponder is serves mock responses
type MockResponder struct {
	doFunc     func(req *http.Request) (*http.Response, error)
	mockData   MockRespList
	lastServed int
	mu         sync.Mutex
}

func defaultDoFunc(req *http.Request) (*http.Response, error) {
	ctxValue := req.Context().Value(contextMockClient)
	if ctxValue == nil {
		panic("no MockResponse context")
	}
	mc, ok := ctxValue.(*MockResponder)
	if !ok {
		panic("returned value is not a MockResponder!")
	}
	log.Printf("mock request url %s", req.URL)
	if mc == nil {
		panic("no data")
	}

	var data MockResp
	length := len(mc.mockData)
	for idx := 0; idx <= length; idx++ {
		if idx == length {
			panic("ran out of data")
		}
		data = mc.mockData[idx]
		if mc.mockData[idx].served {
			continue
		}

		if len(mc.mockData[idx].URL) > 0 {
			m, err := regexp.MatchString(mc.mockData[idx].URL, req.URL.String())
			if err != nil {
				panic("regex pattern issue")
			}
			if !m {
				continue
			}
		}
		mc.mockData[idx].served = true
		mc.lastServed = idx
		break
	}

	// default to 200/OK
	statusCode := data.Code
	if statusCode == 0 {
		statusCode = 200
	}

	resp := &http.Response{
		StatusCode: statusCode,
		Body:       io.NopCloser(bytes.NewReader([]byte(data.Data))),
		Header:     make(http.Header),
	}
	return resp, nil
}

func (m *MockResponder) Do(req *http.Request) (*http.Response, error) {
	// one request at a time!
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.doFunc(req)
}

func (m *MockResponder) SetDoFunc(df func(req *http.Request) (*http.Response, error)) {
	m.doFunc = df
}

func (m *MockResponder) Reset() {
	for _, d := range m.mockData {
		d.served = false
	}
}

func (m *MockResponder) SetData(data MockRespList) {
	m.mockData = data
	m.Reset()
}

func (m *MockResponder) GetData() MockRespList {
	return m.mockData
}

func (m *MockResponder) LastData() []byte {
	return m.mockData[m.lastServed].Data
}

func (m *MockResponder) Empty() bool {
	for _, d := range m.mockData {
		if !d.served {
			return false
		}
	}
	return true
}

func NewMockResponder() (*MockResponder, context.Context) {
	mc := &MockResponder{
		doFunc:   defaultDoFunc,
		mockData: nil,
	}
	return mc, context.WithValue(context.Background(), contextMockClient, mc)
}
