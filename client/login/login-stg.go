package login

import (
	"example.com/micro/client"
	"example.com/micro/config"
	"fmt"
	"gitlab.com/thuocsi.vn-sdk/go-sdk/sdk/common"
)

const (
	loginStg = "POST::/authentication"
)

var (
	stgLoginClient *client.Client
)

func InitLoginStgClient() {
	stgLoginClient = client.NewClient(fmt.Sprintf("%s/core/account/v1", config.StgPublicDomain), 0)
	stgLoginClient.WithConfiguration([]client.Configuration{
		{
			Path: loginStg,
			Name: "stg_login",
		},
	}...)
}

func FuncLoginStg(opts ...client.APIOption) *common.APIResponse {
	o := stgLoginClient.WithAPIOption(opts...)
	if o.Headers == nil || len(o.Headers) == 0 {
		o.Headers["Content-Type"] = "application/json"
		o.Headers["Authorization"] = config.Authorization
	}

	var response *common.APIResponse
	err, _ := stgLoginClient.WithRequest(loginStg, o, &response)
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
