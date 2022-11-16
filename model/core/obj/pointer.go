package obj

import (
	"encoding/json"
	"gitlab.com/thuocsi.vn-sdk/go-sdk/sdk/common"
	"time"
)

func WithTime(i ...time.Time) *time.Time {
	if i == nil {
		_i := time.Now()
		return &_i
	}

	return &i[0]
}

func WithString(i ...string) *string {
	if i == nil {
		_i := ""
		return &_i
	}

	return &i[0]
}

func WithBool(i ...bool) *bool {
	if i == nil {
		_i := false
		return &_i
	}

	return &i[0]
}

func WithInt(i ...int) *int {
	if i == nil {
		_i := 0
		return &_i
	}

	return &i[0]
}

func WithInt64(i ...int64) *int64 {
	if i == nil {
		_i := int64(0)
		return &_i
	}

	return &i[0]
}

func WithFloat(i ...float32) *float32 {
	if i == nil {
		_i := float32(0)
		return &_i
	}

	return &i[0]
}

func WithFloat64(i ...float64) *float64 {
	if i == nil {
		_i := float64(0)
		return &_i
	}

	return &i[0]
}

func WithResponse(content string, template interface{}) error {
	err := json.Unmarshal([]byte(content), &template)
	return err
}

func WithInvalidInput(e error) *common.APIResponse {
	rs := &common.APIResponse{
		Status:    common.APIStatus.Invalid,
		Message:   "Invalid input",
		ErrorCode: "INVALID_INPUT",
	}
	if e != nil {
		rs.Data = []string{e.Error()}
	}
	return rs
}

func WithInvalidSource() *common.APIResponse {
	rs := &common.APIResponse{
		Status:    common.APIStatus.Forbidden,
		Message:   "Invalid account_info",
		ErrorCode: "INVALID_SOURCE",
	}
	return rs
}
