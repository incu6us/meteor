package httputils

import (
	"net/http"
	"time"
	"io/ioutil"
	"io"
)

const (
	// HTTPTimeout - Default HTTP timeout
	HTTPTimeout = 30 * time.Second
)

type httpUtils struct {
	Data     []byte
	Response *http.Response
	Request  *http.Request
	Err      error
	HTTPClient http.Client
}

// set http default timeout
func (h *httpUtils) SetTimeOut(duration time.Duration) {
	if duration == 0 {
		h.HTTPClient.Timeout = HTTPTimeout
		return
	}

	h.HTTPClient.Timeout = duration
}

// get default timeout
func (h *httpUtils) GetTimeOut() time.Duration {
	return h.HTTPClient.Timeout
}

// GetData - get []byte from requested URL with/without authentication
// auth will contains a "username" and "password" fields
func (h *httpUtils) GetData(url string, auth map[string]string) ([]byte, error) {

	// create request
	h.Request, h.Err = http.NewRequest(
		"GET",
		url,
		nil,
	)

	// check if basic auth is needed
	if auth != nil {
		h.Request.SetBasicAuth(auth["username"], auth["password"])
	}

	if h.Response, h.Err = h.HTTPClient.Do(h.Request);
		h.Err != nil {
		//log.Printf("Create request to URL failed: %v", h.Err)
		return nil, h.Err
	}

	defer h.Response.Body.Close()

	// read body
	if h.Data, h.Err = ioutil.ReadAll(h.Response.Body); h.Err != nil {
		//log.Printf("Read from remote URL failed: %v", h.Err)
		return nil, h.Err
	}

	return h.Data, nil
}

func (h *httpUtils) PostData(url string, header map[string]string, body io.Reader, auth map[string]string) ([]byte, error) {

	// create request
	h.Request, h.Err = http.NewRequest(
		"POST",
		url,
		body,
	)

	// set header
	if header != nil {
		for k, v:= range header{
			h.Request.Header.Set(k, v)
		}
	}

	// check if basic auth is needed
	if auth != nil {
		h.Request.SetBasicAuth(auth["username"], auth["password"])
	}

	if h.Response, h.Err = h.HTTPClient.Do(h.Request);
		h.Err != nil {
		//log.Printf("Create request to URL failed: %v", h.Err)
		return nil, h.Err
	}

	defer h.Response.Body.Close()

	// read body
	if h.Data, h.Err = ioutil.ReadAll(h.Response.Body); h.Err != nil {
		//log.Printf("Read from remote URL failed: %v", h.Err)
		return nil, h.Err
	}

	return h.Data, nil
}

// HTTPUtils - main methods
type HTTPUtils interface {
	//SetTimeOut(time.Duration)
	//GetTimeOut() time.Duration
	GetData(string, map[string]string) ([]byte, error)
	PostData(url string, header map[string]string, body io.Reader, auth map[string]string) ([]byte, error)
}

// NewHTTPUtil - main instance of the util
func NewHTTPUtil() HTTPUtils {
	util := &httpUtils{}
	util.SetTimeOut(0)

	//log.Printf("Set default HTTP timeout: %s", util.GetTimeOut())

	return util
}
