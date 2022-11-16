package sdk

import (
	"encoding/json"
	"gitlab.com/thuocsi.vn-sdk/go-sdk/sdk/common"
	"strings"

	"gitlab.com/thuocsi.vn-sdk/go-sdk/sdk/thriftapi"
)

// APIThriftRequest ...
type APIThriftRequest struct {
	t          string
	context    *thriftapi.APIRequest
	attributes map[string]interface{}
	variables map[string]string
}

func newThriftAPIRequest(e *thriftapi.APIRequest) APIRequest {
	return &APIThriftRequest{
		t:          "THRIFT",
		context:    e,
		attributes: make(map[string]interface{}),
		variables: map[string]string{},
	}
}

func (req *APIThriftRequest) GetPath() string {
	return req.context.GetPath()
}

func (req *APIThriftRequest) GetIP() string {
	forwarded := req.GetHeader("X-Forwarded-For")
	if forwarded == "" {
		return ""
	}

	splitted := strings.Split(forwarded, ",")
	return splitted[0]
}

func (req *APIThriftRequest) GetMethod() *common.MethodValue {
	var s = req.context.GetMethod()
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

// GetParam ...
func (req *APIThriftRequest) GetParam(name string) string {
	params := req.context.GetParams()
	if params == nil {
		return ""
	}
	return params[name]
}

// GetParams ...
func (req *APIThriftRequest) GetParams() map[string]string {
	return req.context.GetParams()
}

// GetContent ...
func (req *APIThriftRequest) GetContent(data interface{}) error {
	return json.Unmarshal([]byte(req.context.Content), &data)
}

// GetContentText ...
func (req *APIThriftRequest) GetContentText() string {
	return req.context.Content
}

// GetHeader ...
func (req *APIThriftRequest) GetHeader(name string) string {
	headers := req.context.GetHeaders()
	if headers == nil {
		return ""
	}
	return headers[name]
}

// GetHeaders ...
func (req *APIThriftRequest) GetHeaders() map[string]string {
	return req.context.GetHeaders()
}

// GetAttribute ...
func (req *APIThriftRequest) GetAttribute(name string) interface{} {
	return req.attributes[name]
}

// SetAttribute ...
func (req *APIThriftRequest) SetAttribute(name string, value interface{}) {
	req.attributes[name] = value
}

// GetAttr ...
func (req *APIThriftRequest) GetAttr(name string) interface{} {
	return req.attributes[name]
}

// SetAttr ...
func (req *APIThriftRequest) SetAttr(name string, value interface{}) {
	req.attributes[name] = value
}

// GetAttr ...
func (req *APIThriftRequest) GetVar(name string) string {
	return req.variables[name]
}

// SetAttr ...
func (req *APIThriftRequest) SetVar(name string, value string) {
	req.variables[name] = value
}
