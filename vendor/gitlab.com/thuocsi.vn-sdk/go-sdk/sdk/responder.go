package sdk

import (
	"errors"
	"fmt"
	"gitlab.com/thuocsi.vn-sdk/go-sdk/sdk/common"
	"net/http"
	"reflect"
	"time"

	"github.com/labstack/echo"
	"gitlab.com/thuocsi.vn-sdk/go-sdk/sdk/thriftapi"
)

// APIResponder ...
type APIResponder interface {
	Respond(*common.APIResponse) error
	GetThriftResponse() *thriftapi.APIResponse
}

// HTTPAPIResponder This is response object with JSON format
type HTTPAPIResponder struct {
	t        string
	context  echo.Context
	start    time.Time
	hostname string
}

func newHTTPAPIResponder(c echo.Context, hostname string) APIResponder {
	return &HTTPAPIResponder{
		t:        "HTTP",
		start:    time.Now(),
		context:  c,
		hostname: hostname,
	}
}

// Respond ..
func (resp *HTTPAPIResponder) Respond(response *common.APIResponse) error {
	var context = resp.context

	if response.Data != nil && reflect.TypeOf(response.Data).Kind() != reflect.Slice {
		return errors.New("data response must be a slice")
	}

	if response.Headers != nil {
		header := context.Response().Header()
		for key, value := range response.Headers {
			header.Set(key, value)
		}
		response.Headers = nil
	}

	var dif = float64(time.Since(resp.start).Nanoseconds()) / 1000000
	context.Response().Header().Set("X-Execution-Time", fmt.Sprintf("%.4f ms", dif))
	context.Response().Header().Set("X-Hostname", resp.hostname)

	switch response.Status {
	case common.APIStatus.Ok:
		return context.JSON(http.StatusOK, response)
	case common.APIStatus.Error:
		return context.JSON(http.StatusInternalServerError, response)
	case common.APIStatus.Forbidden:
		return context.JSON(http.StatusForbidden, response)
	case common.APIStatus.Invalid:
		return context.JSON(http.StatusBadRequest, response)
	case common.APIStatus.NotFound:
		return context.JSON(http.StatusNotFound, response)
	case common.APIStatus.Unauthorized:
		return context.JSON(http.StatusUnauthorized, response)
	case common.APIStatus.Existed:
		return context.JSON(http.StatusConflict, response)
	}

	return context.JSON(http.StatusBadRequest, response)
}

// GetThriftResponse ..
func (resp *HTTPAPIResponder) GetThriftResponse() *thriftapi.APIResponse {
	return nil
}
