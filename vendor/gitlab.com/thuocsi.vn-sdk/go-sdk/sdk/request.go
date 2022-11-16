package sdk

import (
	"encoding/json"
	"gitlab.com/thuocsi.vn-sdk/go-sdk/sdk/common"
	"io/ioutil"
	"strings"

	"github.com/labstack/echo"
)

// APIRequest ...
type APIRequest interface {
	GetPath() string
	GetMethod() *common.MethodValue
	GetParam(string) string
	GetParams() map[string]string
	GetHeader(string) string
	GetHeaders() map[string]string
	GetContent(interface{}) error
	GetContentText() string
	GetAttribute(string) interface{}
	SetAttribute(string, interface{})
	SetVar(string, string)
	GetVar(string) string
	GetIP() string
}

// HTTPAPIRequest This is response object with JSON format
type HTTPAPIRequest struct {
	t       string
	context echo.Context
	body    string
}

func newHTTPAPIRequest(e echo.Context) APIRequest {
	return &HTTPAPIRequest{
		t:       "HTTP",
		context: e,
	}
}

func (req *HTTPAPIRequest) GetPath() string {
	return req.context.Path()
}

func (req *HTTPAPIRequest) GetMethod() *common.MethodValue {
	var s = req.context.Request().Method
	switch s {
	case "GET":
		return common.APIMethod.GET
	case "POST":
		return common.APIMethod.POST
	case "PUT":
		return common.APIMethod.PUT
	case "DELETE":
		return common.APIMethod.DELETE
	}

	return &common.MethodValue{Value: s}
}

// GetVar ...
func (req *HTTPAPIRequest) GetVar(name string) string {
	return req.context.Param(name)
}

// GetVar ...
func (req *HTTPAPIRequest) SetVar(name string, value string) {

}

// GetParam ...
func (req *HTTPAPIRequest) GetParam(name string) string {
	return req.context.QueryParam(name)
}

// GetParams ...
func (req *HTTPAPIRequest) GetParams() map[string]string {
	var vals = req.context.QueryParams()
	var m = make(map[string]string)
	for key := range vals {
		m[key] = vals.Get(key)
	}
	return m
}

// GetContent ...
func (req *HTTPAPIRequest) GetContent(data interface{}) error {

	return json.Unmarshal([]byte(req.GetContentText()), data)
}

// GetContentText ...
func (req *HTTPAPIRequest) GetContentText() string {
	if req.body == "" {
		var bodyBytes []byte
		if req.context.Request().Body != nil {
			bodyBytes, _ = ioutil.ReadAll(req.context.Request().Body)
		}

		req.body = string(bodyBytes)
	}

	return req.body
}

// GetHeader ...
func (req *HTTPAPIRequest) GetHeader(name string) string {
	return req.context.Request().Header.Get(name)
}

// GetHeaders ...
func (req *HTTPAPIRequest) GetHeaders() map[string]string {
	var vals = req.context.Request().Header
	var m = make(map[string]string)
	for key := range vals {
		m[key] = vals.Get(key)
	}
	return m
}

// GetAttribute ...
func (req *HTTPAPIRequest) GetAttribute(name string) interface{} {
	return req.context.Get(name)
}

// SetAttribute ...
func (req *HTTPAPIRequest) SetAttribute(name string, value interface{}) {
	req.context.Set(name, value)
}

func (req *HTTPAPIRequest) GetIP() string {
	// for forwarded case
	forwarded := req.GetHeader("X-Forwarded-For")
	if forwarded == "" {
		httpReq := req.context.Request()
		return strings.Split(httpReq.RemoteAddr, ":")[0]
	}

	splitted := strings.Split(forwarded, ",")
	return splitted[0]
}
