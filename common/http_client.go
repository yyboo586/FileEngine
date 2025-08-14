package common

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

type HTTPClient interface {
	GET(ctx context.Context, url string, header map[string]interface{}) (resp interface{}, err error)
	POST(ctx context.Context, url string, header map[string]interface{}, body interface{}) (resp interface{}, err error)
}

type httpClient struct {
	client *http.Client
}

var (
	cOnce sync.Once
	c     *httpClient
)

func NewHTTPClient() HTTPClient {
	cOnce.Do(func() {
		c = &httpClient{
			client: &http.Client{
				Timeout: time.Second * 10,
				CheckRedirect: func(req *http.Request, via []*http.Request) error {
					return nil
				},
			},
		}
	})
	return c
}

func (c *httpClient) GET(ctx context.Context, url string, header map[string]interface{}) (resp interface{}, err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return
	}

	c.setHeader(req, header)

	return c.do(req)
}

func (c *httpClient) POST(ctx context.Context, url string, header map[string]interface{}, body interface{}) (resp interface{}, err error) {
	var dataByte []byte
	switch data := body.(type) {
	case []byte:
		dataByte = data
	case string:
		dataByte = []byte(data)
	default:
		dataByte, err = json.Marshal(data)
		if err != nil {
			return
		}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(dataByte))
	if err != nil {
		return
	}
	c.setHeader(req, header)

	resp, err = c.do(req)
	if err != nil {
		return
	}

	return
}

func (c *httpClient) do(req *http.Request) (resp interface{}, err error) {
	response, err := c.client.Do(req)
	if err != nil {
		return
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http request failed, status code: %d", response.StatusCode)
	}

	respByte, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body failed, err: %w", err)
	}

	err = json.Unmarshal(respByte, &resp)
	if err != nil {
		return nil, fmt.Errorf("unmarshal response body failed, err: %w", err)
	}

	log.Println(resp)

	return
}

func (c *httpClient) setHeader(req *http.Request, header map[string]interface{}) {
	for k, v := range header {
		req.Header.Set(k, v.(string))
	}
}
