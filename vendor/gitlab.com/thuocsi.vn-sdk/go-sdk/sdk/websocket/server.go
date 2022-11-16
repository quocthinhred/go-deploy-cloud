package websocket

import (
	"fmt"
	"golang.org/x/net/websocket"
	"net/http"
	"strconv"
	"sync"
	"time"
)

//WSServer represent web socket server
type WSServer struct {
	Port int
	Name string

	server    *http.Server
	mux       *http.ServeMux
	lock      *sync.Mutex
	idCounter int
	closedCon int
	routing   map[string]*wsRoute
	Timeout   int
}

// genConId generate new connection id
func (wss *WSServer) genConId() int {
	wss.lock.Lock()
	defer wss.lock.Unlock()

	wss.idCounter++
	return wss.idCounter
}

// closeACon close one ws connection
func (wss *WSServer) closeACon() {
	wss.lock.Lock()
	defer wss.lock.Unlock()

	wss.closedCon++
}

//NewWebSocketServer create new WS server
func NewWebSocketServer(name string) (wss *WSServer) {
	wss = &WSServer{
		Name:    name,
		lock:    &sync.Mutex{},
		routing: map[string]*wsRoute{},
		mux:     &http.ServeMux{},
	}

	return wss
}

//NewRoute create new coordinator, which can routing WS request by path
func (wss *WSServer) NewRoute(path string) *wsRoute {
	wsr := newWSRoute()
	wss.routing[path] = wsr

	// bind this coordinator to path
	wss.mux.Handle(path, websocket.Handler(func(conn *websocket.Conn) {
		con := newConnection(wss.genConId(), conn, wss.Timeout)
		wsr.addCon(con)

		// setup handler
		if wsr.OnConnected != nil {
			wsr.OnConnected(con)
		}

		// chatting with client
		wsr.transferring(con)
	}))

	return wsr
}

//GetRoute get coordinator of given path
func (wss *WSServer) GetRoute(path string) *wsRoute {
	return wss.routing[path]
}

//Expose expose port
func (wss *WSServer) Expose(port int) {
	wss.Port = port
}

//GetActiveCon get active connection number in server
func (wss *WSServer) GetActiveCon() int {
	wss.lock.Lock()
	defer wss.lock.Unlock()

	return wss.idCounter - wss.closedCon
}

//Start Start API server
func (wss *WSServer) Start() {
	// prevent Start run twice
	if wss.server != nil {
		return
	}

	// setup http server
	wss.server = &http.Server{
		Addr:         ":" + strconv.Itoa(wss.Port),
		Handler:      wss.mux,
		ReadTimeout:  1800 * time.Second,
		WriteTimeout: 1800 * time.Second,
	}

	//start to listen
	fmt.Println("  [ WS Server " + wss.Name + " ] Try to listen at " + strconv.Itoa(wss.Port))
	err := wss.server.ListenAndServe()
	if err != nil {
		panic(err)
	}
}
