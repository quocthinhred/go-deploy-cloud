package sdk

import (
	"encoding/json"
	"gitlab.com/thuocsi.vn-sdk/go-sdk/sdk/common"
)

// OutboundAPIRequest Request to call other service
type OutboundAPIRequest struct {
	Method  string            `json:"method" bson:"method"`
	Path    string            `json:"path" bson:"path"`
	Params  map[string]string `json:"params,omitempty" bson:"params,omitempty"`
	Headers map[string]string `json:"headers,headers" bson:"headers,omitempty"`
	Content string            `json:"content,omitempty" bson:"content,omitempty"`
}

func NewOutboundAPIRequest(method string, path string, params map[string]string, content string, headers map[string]string) APIRequest {
	return &OutboundAPIRequest{
		Method:  method,
		Path:    path,
		Params:  params,
		Content: content,
		Headers: headers,
	}
}

// GetPath ..
func (req *OutboundAPIRequest) GetPath() string {
	return req.Path
}

// GetPath ..
func (req *OutboundAPIRequest) GetIP() string {
	return "GetIP() not implemented"
}

// GetMethod ..
func (req *OutboundAPIRequest) GetMethod() *common.MethodValue {
	var s = req.Method
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
func (req *OutboundAPIRequest) GetVar(name string) string {
	return req.Params[name]
}

// GetParam ...
func (req *OutboundAPIRequest) GetParam(name string) string {
	return req.Params[name]
}

// GetParams ...
func (req *OutboundAPIRequest) GetParams() map[string]string {
	return req.Params
}

// GetContent ...
func (req *OutboundAPIRequest) GetContent(data interface{}) error {
	json.Unmarshal([]byte(req.Content), &data)
	return nil
}

// GetContentText ...
func (req *OutboundAPIRequest) GetContentText() string {
	return req.Content
}

// GetHeader ...
func (req *OutboundAPIRequest) GetHeader(name string) string {
	return req.Headers[name]
}

// GetHeaders ...
func (req *OutboundAPIRequest) GetHeaders() map[string]string {
	return req.Headers
}

// GetAttribute ...
func (req *OutboundAPIRequest) GetAttribute(name string) interface{} {
	return nil
}

// SetAttribute ...
func (req *OutboundAPIRequest) SetAttribute(name string, value interface{}) {

}

// GetAttr ...
func (req *OutboundAPIRequest) GetAttr(name string) interface{} {
	return nil
}

// SetAttr ...
func (req *OutboundAPIRequest) SetAttr(name string, value interface{}) {
	// do nothing
}


// SetAttr ...
func (req *OutboundAPIRequest) SetVar(name string, value string) {
	// do nothing
}