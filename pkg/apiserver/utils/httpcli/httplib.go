package httpcli

import (
	"bytes"
	"crypto/tls"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// 默认http连接时间
var defaultSettings = HTTPSettings{
	ConnectTimeout:   60 * time.Second,
	ReadWriteTimeout: 60 * time.Second,
}

type HTTPSettings struct {
	ConnectTimeout   time.Duration
	ReadWriteTimeout time.Duration
	Transport        http.RoundTripper
	TLSClientConfig  *tls.Config
	Proxy            func(r *http.Request) (*url.URL, error)
	Retries          int
}

type HttpRequest struct {
	url      string
	request  *http.Request
	params   map[string][]string
	settings HTTPSettings
	response *http.Response

	body []byte
}

func NewHttpRequest(rawUrl, method string) *HttpRequest {
	var response http.Response

	u, err := url.Parse(rawUrl)
	if err != nil {
		log.Println("httplib:", err)
	}

	request := http.Request{
		URL:        u,
		Method:     method,
		Header:     make(http.Header),
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
	}

	return &HttpRequest{
		url:      rawUrl,
		request:  &request,
		params:   map[string][]string{},
		settings: defaultSettings,
		response: &response,
	}
}

func Get(url string) *HttpRequest {
	return NewHttpRequest(url, "GET")
}

func Patch(url string) *HttpRequest {
	return NewHttpRequest(url, "PATCH")
}

func Post(url string) *HttpRequest {
	return NewHttpRequest(url, "Post")
}

func Put(url string) *HttpRequest {
	return NewHttpRequest(url, "Put")
}

func Delete(url string) *HttpRequest {
	return NewHttpRequest(url, "DELETE")
}

func (h *HttpRequest) buildURL(paramBody string) {
	if h.request.Method == "GET" && len(paramBody) > 0 {
		if strings.Contains(h.url, "?") {
			h.url += "&" + paramBody
		} else {
			h.url += "?" + paramBody
		}
		return
	}

	if h.request.Method == "POST" || h.request.Method == "PUT" || h.request.Method == "DELETE" && h.request.Body == nil {
		if len(paramBody) > 0 {
			h.AddHeader("Content-Type", "application/x-www-form-urlencoded")
			h.Body(paramBody)
		}
	}
}

func (h *HttpRequest) AddHeader(key, value string) {
	h.request.Header.Set(key, value)
}

func (h *HttpRequest) Body(data interface{}) {
	switch t := data.(type) {
	case string:
		buffer := bytes.NewBufferString(t)
		h.request.Body = ioutil.NopCloser(buffer)
		h.request.ContentLength = int64(len(t))
	case []byte:
		buffer := bytes.NewBuffer(t)
		h.request.Body = ioutil.NopCloser(buffer)
		h.request.ContentLength = int64(len(t))
	}
}

func (h *HttpRequest) doRequest() (response *http.Response, err error) {
	var paramBody string
	if len(h.params) > 0 {
		var buf bytes.Buffer
		for k, v := range h.params {
			for _, vv := range v {
				buf.WriteString(url.QueryEscape(k))
				buf.WriteByte('=')
				buf.WriteString(url.QueryEscape(vv))
				buf.WriteByte('&')
			}
		}
		paramBody = buf.String()
		paramBody = paramBody[0 : len(paramBody)-1]
	}

	h.buildURL(paramBody)

	urlParsed, err := url.Parse(h.url)
	if err != nil {
		return nil, err
	}
	h.request.URL = urlParsed

	trans := h.settings.Transport
	if trans == nil {
		// 创建默认的transport
		trans = &http.Transport{
			TLSClientConfig:     h.settings.TLSClientConfig,
			Proxy:               h.settings.Proxy,
			Dial:                TimeoutDialer(h.settings.ConnectTimeout, h.settings.ReadWriteTimeout),
			MaxIdleConnsPerHost: 100,
		}
	} else {
		// 更新transport

	}

	client := &http.Client{
		Transport: trans,
	}

	// Retries默认为0，请求只执行一次
	// Retries为-1，表示一直执行直到成功为止
	// Retries为其他值，表示重试次数
	for i := 0; h.settings.Retries == -1 || i <= h.settings.Retries; i++ {
		response, err = client.Do(h.request)
		if err == nil {
			break
		}
	}

	h.response = response
	return response, err
}

func TimeoutDialer(cTimeout time.Duration, rwTimeout time.Duration) func(net, addr string) (c net.Conn, err error) {
	return func(netw, addr string) (c net.Conn, err error) {
		conn, err := net.DialTimeout(netw, addr, cTimeout)
		if err != nil {
			return nil, err
		}
		err = conn.SetDeadline(time.Now().Add(rwTimeout))
		return conn, err
	}
}

func (h *HttpRequest) Send() ([]byte, error) {
	response, err := h.doRequest()
	if err != nil {
		return nil, err
	}
	// 无返回值
	if response.Body == nil {
		return nil, nil
	}
	defer response.Body.Close()

	return ioutil.ReadAll(response.Body)
}
