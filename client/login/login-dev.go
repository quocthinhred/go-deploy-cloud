package login

import (
	"example.com/micro/client"
	"example.com/micro/config"
	"fmt"
	"gitlab.com/thuocsi.vn-sdk/go-sdk/sdk/common"
)

const (
	loginDev = "POST::/authentication"
)

var (
	devLoginClient *client.Client
)

func InitLoginDevClient() {
	devLoginClient = client.NewClient(fmt.Sprintf("%s/core/account/v1", config.DevPublicDomain), 0)
	devLoginClient.WithConfiguration([]client.Configuration{
		{
			Path: loginDev,
			Name: "dev_login",
		},
	}...)
}

func FuncLoginDev(opts ...client.APIOption) *common.APIResponse {
	o := devLoginClient.WithAPIOption(opts...)
	if o.Headers == nil || len(o.Headers) == 0 {
		o.Headers["Content-Type"] = "application/json"
		o.Headers["Authorization"] = config.Authorization
	}

	var response *common.APIResponse
	err, _ := devLoginClient.WithRequest(loginDev, o, &response)
	if err == client.ErrConfiguration {
		return &common.APIResponse{
			Status:    common.APIStatus.Error,
			Message:   "Invalid endpoint configuration",
			ErrorCode: "INVALID_ENDPOINT_CONFIGURATION",
		}
	}

	if response == nil {
		return &common.APIResponse{
			Status:    common.APIStatus.Error,
			Message:   "No response",
			ErrorCode: "N0_RESPONSE",
		}
	}

	if err != nil {
		return &common.APIResponse{
			Status:  common.APIStatus.Error,
			Message: err.Error(),
		}
	}

	return &common.APIResponse{
		Status:    response.Status,
		Message:   response.Message,
		ErrorCode: response.ErrorCode,
		Data:      response.Data,
		Total:     response.Total,
	}
}
