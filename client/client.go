package client

import (
	"encoding/json"
	"errors"
	"example.com/micro/model/core/obj"
	sdk_client "gitlab.com/thuocsi.vn-sdk/go-sdk/sdk/client"
	"gitlab.com/thuocsi.vn-sdk/go-sdk/sdk/common"
	"go.mongodb.org/mongo-driver/mongo"
	"strconv"
	"strings"
	"time"
)

const (
	defaultTimeout = 30 * time.Second
)

var (
	ErrConfiguration = errors.New("invalid endpoint configuration")
)

type APIOption struct {
	Keys    []string // for query log client (if not disable SaveLog)
	SaveLog *bool    // disable logging if SaveLog = false

	URL      string
	Headers  map[string]string
	Params   map[string]string
	Vars     map[string]string
	GetTotal *bool       // will be merged into Params
	Offset   *int        // will be merged into Params
	Limit    *int        // will be merged into Params
	Q        interface{} // will be merged into Params; just accept string or obj / map
	Body     interface{}
}

func getAPIOption(opts ...APIOption) APIOption {
	o := APIOption{
		Keys:    make([]string, 0),
		SaveLog: nil,
		Headers: make(map[string]string),
		Params:  make(map[string]string),
		Vars:    make(map[string]string),
		Body:    nil,
	}

	for _, opt := range opts {
		if opt.Keys != nil {
			o.Keys = append(o.Keys, opt.Keys...)
		}

		if opt.SaveLog != nil {
			o.SaveLog = opt.SaveLog
		}

		if opt.Params != nil {
			o.Params = opt.Params
		}

		if opt.Vars != nil {
			o.Vars = opt.Vars
		}

		if opt.Q != nil {
			if s, ok := opt.Q.(string); ok {
				o.Params["q"] = s
			} else {
				b, _ := json.Marshal(opt.Q)
				o.Params["q"] = string(b)
			}
		}

		if opt.GetTotal != nil {
			o.Params["getTotal"] = strconv.FormatBool(*opt.GetTotal)
		}

		if opt.Offset != nil {
			o.Params["offset"] = strconv.Itoa(*opt.Offset)
		}

		if opt.Limit != nil {
			o.Params["limit"] = strconv.Itoa(*opt.Limit)
		}

		if opt.URL != "" {
			o.URL = opt.URL
		}

		if opt.Body != nil {
			o.Body = opt.Body
		}

		if opt.Headers != nil {
			o.Headers = opt.Headers
		}
	}

	return o
}

type Configuration struct {
	Path     string
	Name     string
	Database *mongo.Database
	Timeout  time.Duration
}

type Client struct {
	database         *mongo.Database
	host             string
	clientWithoutLog *sdk_client.RestClient
	clients          map[string]*sdk_client.RestClient
	headers          map[string]string
}

func NewClient(host string, timeout time.Duration) *Client {
	if timeout < 0 {
		timeout = defaultTimeout
	}

	c := &Client{
		host:    host,
		clients: make(map[string]*sdk_client.RestClient),
	}
	c.clientWithoutLog = sdk_client.NewRESTClient(host, "", timeout, 0, 0)
	c.clientWithoutLog.AcceptHTTPError(true)
	return c
}

func (c *Client) WithDatabase(d ...*mongo.Database) {
	if d != nil {
		c.database = d[0]
	}
}

func (c *Client) WithConfiguration(conf ...Configuration) {
	for _, config := range conf {
		if config.Timeout < 0 {
			config.Timeout = defaultTimeout
		}
		cl := sdk_client.NewRESTClient(c.host, config.Name, config.Timeout, 0, 0)
		if config.Database != nil {
			cl.SetDBLog(config.Database)
		} else if c.database != nil {
			cl.SetDBLog(c.database)
		}
		cl.AcceptHTTPError(true)

		c.clients[config.Path] = cl
	}
}

func (c Client) WithAPI(api ...string) *sdk_client.RestClient {
	if api == nil {
		return c.clientWithoutLog
	}

	if cl, ok := c.clients[api[0]]; !ok || cl == nil {
		return c.clientWithoutLog
	} else {
		return cl
	}
}

func (c Client) WithMethod(method ...string) sdk_client.HTTPMethod {
	if method == nil {
		return sdk_client.HTTPMethods.Head
	}

	switch method[0] {
	default:
		return sdk_client.HTTPMethods.Head
	case "GET":
		return sdk_client.HTTPMethods.Get
	case "POST":
		return sdk_client.HTTPMethods.Post
	case "PUT":
		return sdk_client.HTTPMethods.Put
	case "DELETE":
		return sdk_client.HTTPMethods.Delete
	}
}

func (c Client) WithAPIOption(opts ...APIOption) APIOption {
	return getAPIOption(opts...)
}

func (c Client) WithStatus(code int) string {
	var r = code / 100
	switch r {
	case 2:
		return common.APIStatus.Ok
	case 4:
		switch code {
		case 400:
			return common.APIStatus.Invalid
		case 401:
			return common.APIStatus.Unauthorized
		case 403:
			return common.APIStatus.Forbidden
		case 404:
			return common.APIStatus.NotFound
		case 409:
			return common.APIStatus.Existed
		}
	case 5:
		return common.APIStatus.Error
	}

	return "does not have the implementation"
}

// WithRequest ...
/**
 * @param api <method>::</path>
 * @param o combined API options by WithAPIOption
 * @param template for data type conversion in response
 */
func (c Client) WithRequest(api string, o APIOption, template interface{}) (error, int) {
	cl := c.WithAPI(api)
	if o.SaveLog != nil && !*o.SaveLog {
		cl = c.clientWithoutLog
	}

	path := string(api)
	if o.Vars != nil {
		for k, v := range o.Vars {
			path = strings.ReplaceAll(path, ":var_"+k, v)
		}
	}

	parts := strings.Split(path, "::")
	if len(parts) != 2 {
		return ErrConfiguration, 0
	}
	m := c.WithMethod(parts[0])
	result, err := cl.MakeHTTPRequestWithKey(m, o.Headers, o.Params, o.Body, parts[1], &o.Keys)
	if err != nil {
		return err, result.Code
	}

	return obj.WithResponse(result.Body, template), result.Code
}
