package sdk

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"

	"gitlab.com/thuocsi.vn-sdk/go-sdk/sdk/common"

	"gitlab.com/thuocsi.vn-sdk/go-sdk/sdk/thriftapi"

	"github.com/apache/thrift/lib/go/thrift"
)

// ThriftServer ...
type ThriftServer struct {
	rootServer    *thrift.TSimpleServer
	thriftHandler *ThriftHandler
	port          int
	ID            int
	hostname      string
}

// NewThriftServer ...
func newThriftServer(id int, hostname string) APIServer {
	return &ThriftServer{
		thriftHandler: &ThriftHandler{
			Handlers: make(map[string]Handler),
			hostname: hostname,
		},
		ID:       id,
		port:     9090,
		hostname: hostname,
	}
}

// SetHandle ...
func (server *ThriftServer) SetHandler(method *common.MethodValue, path string, fn Handler) error {
	fullPath := string(method.Value) + "://" + path
	server.thriftHandler.Handlers[fullPath] = fn
	return nil
}

//PreRequest ...
func (server *ThriftServer) PreRequest(fn PreHandler) error {
	server.thriftHandler.preHandler = fn
	return nil
}

//Expose Add api handler
func (server *ThriftServer) Expose(port int) {
	server.port = port
}

//Start Start API server
func (server *ThriftServer) Start(wg *sync.WaitGroup) {
	var ps = strconv.Itoa(server.port)
	fmt.Println("  [ Thrift API Server " + strconv.Itoa(server.ID) + " ] Try to listen at " + ps)

	var transport thrift.TServerTransport
	transport, _ = thrift.NewTServerSocket("0.0.0.0:" + ps)
	proc := thriftapi.NewAPIServiceProcessor(server.thriftHandler)
	server.rootServer = thrift.NewTSimpleServer4(proc, transport,
		thrift.NewTFramedTransportFactory(thrift.NewTBufferedTransportFactory(24*1024)),
		thrift.NewTBinaryProtocolFactoryDefault())
	server.rootServer.Serve()
	wg.Done()
}

func (server *ThriftServer) GetHostname() string {
	return server.hostname
}

// ServeHTTP ...
func (server *ThriftServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	return
}

// ThriftHandler ...
type ThriftHandler struct {
	Handlers   map[string]Handler
	preHandler PreHandler
	hostname   string
}

// Call Override abstract thrift interface
func (th *ThriftHandler) Call(ctx context.Context, request *thriftapi.APIRequest) (r *thriftapi.APIResponse, err error) {
	defer func(){
		if rec := recover(); rec != nil {
			r = &thriftapi.APIResponse{
				Status:  thriftapi.Status_ERROR,
				Message: "There is an error, please try again later.",
				ErrorCode: "INTERNAL_SERVICE_ERROR",
			}
			log.Println("panic: ",rec, string(debug.Stack()))
		}
	}()

	var req = newThriftAPIRequest(request)
	var responder = newThriftAPIResponder(th.hostname)
	var resp *thriftapi.APIResponse

	// process pre-request
	if th.preHandler != nil {
		err := th.preHandler(req, responder)
		resp = responder.GetThriftResponse()
		if err != nil || resp != nil {
			if resp == nil {
				resp = &thriftapi.APIResponse{
					Status:  thriftapi.Status_ERROR,
					Message: "PreRequest error: " + err.Error(),
				}
			}
			return resp, err
		}
	}

	// process routing
	method := req.GetMethod()
	path := request.GetPath()
	fullPath := method.Value + "://" + path
	if th.Handlers[fullPath] != nil {
		th.Handlers[fullPath](req, responder)
		return responder.GetThriftResponse(), nil
	} else {

		inputParts := strings.Split(path, "/")

		// setup data for selected handler
		var selectedHandler Handler = nil
		var selectedScore = 0
		var varMap = map[string]string{}

		for full, hdl := range th.Handlers {

			// init for each handler
			var score = 0
			var tempMap = map[string]string{}

			methodPath := strings.Split(full, "://")
			if method.Value != methodPath[0] {
				continue
			}

			// scan path parts
			validation := true
			pathParts := strings.Split(methodPath[1], "/")
			for i, part := range pathParts {
				if i < len(inputParts) {
					if strings.HasPrefix(part, ":") {
						tempMap[part[1:]] = inputParts[i]
					} else if part != inputParts[i] {
						validation = false
						break
					}
					score = i + 1 // if match at parts[0] => score = 1
				} else {
					break
				}
			}

			if !validation {
				continue
			}

			// if this handler has higher score
			if score > selectedScore || (score == selectedScore && len(pathParts) == len(inputParts)){ // match length has higher priority
				varMap = tempMap
				selectedHandler = hdl
				selectedScore = score
			}

		}

		if selectedHandler != nil {
			for key, value := range varMap {
				req.SetVar(key, value)
			}
			selectedHandler(req, responder)
			return responder.GetThriftResponse(), nil
		}
	}

	return &thriftapi.APIResponse{
		Status:  thriftapi.Status_NOT_FOUND,
		Message: "API Method/Path " + method.Value + " " + path + " isn't found",
		ErrorCode: "API_NOT_FOUND",
	}, nil
}
