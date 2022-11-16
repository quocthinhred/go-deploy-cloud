package client

import (
	"gitlab.com/thuocsi.vn-sdk/go-sdk/sdk"
	"gitlab.com/thuocsi.vn-sdk/go-sdk/sdk/common"
	"time"
)

// APIClient
type APIClient interface {
	MakeRequest(sdk.APIRequest) *common.APIResponse
	SetDebug(val bool)
}

// APIClientConfiguration
type APIClientConfiguration struct {
	Address       string
	Protocol      string
	Timeout       time.Duration
	MaxRetry      int
	WaitToRetry   time.Duration
	LoggingCol    string
	LogExpiration *time.Duration
	MaxConnection int
	ErrorLogOnly  bool
}

func NewAPIClient(config *APIClientConfiguration) APIClient {
	switch config.Protocol {
	case "THRIFT":
		return NewThriftClient(config.Address, config.Timeout, config.MaxConnection, config.MaxRetry, config.WaitToRetry, 600)
	case "HTTP":
		return NewHTTPClient(config)
	}
	return nil
}
