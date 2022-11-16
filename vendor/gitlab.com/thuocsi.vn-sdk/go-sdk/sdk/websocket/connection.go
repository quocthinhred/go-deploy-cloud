package websocket

import (
	"errors"
	"golang.org/x/net/websocket"
	"strings"
	"sync"
	"time"
)

// Connection represent a ws connection
type Connection struct {
	Id       int
	Attached map[string]interface{}
	Key      string

	rootCon  *websocket.Conn
	isActive bool
	lock     *sync.Mutex

	timeout      int
	sentMsgCount int
	readMsgCount int
	lastWrite    *time.Time
	lastRead     *time.Time
}

func newConnection(id int, conn *websocket.Conn, timeout int) *Connection {
	con := &Connection{
		Id:       id,
		rootCon:  conn,
		Attached: make(map[string]interface{}),
		isActive: true,
		lock:     &sync.Mutex{},
		timeout:  timeout,
	}
	return con
}

// getRooCon return root
func (con *Connection) getRootCon() *websocket.Conn {
	return con.rootCon
}

// Send push message to client side
func (con *Connection) Send(message string) error {
	con.lock.Lock()
	if con.timeout > 0 {
		con.rootCon.SetWriteDeadline(time.Now().Add(time.Duration(con.timeout) * time.Second))
	}
	n, err := con.rootCon.Write([]byte(message))
	if err == nil && n > 0 {
		con.sentMsgCount++
	}
	con.lock.Unlock()
	return err
}

// Read push message to client side
func (con *Connection) Read() (message string, err error) {
	defer func(){
		if rec := recover(); rec != nil {
			message = ""
			err = errors.New("panic & recovered")
		}
	}()

	var msg = make([]byte, con.rootCon.MaxPayloadBytes)
	n, err := con.rootCon.Read(msg)
	if err == nil && n > 0 {
		con.lock.Lock()
		con.readMsgCount++
		con.lock.Unlock()
	}
	return string(msg[:n]), err
}

// IsActive check status of connection
func (con *Connection) IsActive() bool {
	return con.isActive
}

// CLose close the connection
func (con *Connection) Close() {
	if con.isActive {
		con.rootCon.Close()
		con.isActive = false
	}
}

// Deactive deactive status of connection
func (con *Connection) Deactive() {
	con.isActive = false
}

func (con *Connection) GetIP() string {
	req := con.rootCon.Request()
	forwarded := req.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		if strings.Contains(forwarded, ","){
			forwarded = strings.Split(forwarded, ",")[0]
		}
		return strings.Trim(forwarded, " ")
	}

	adr := con.rootCon.Request().RemoteAddr
	n := strings.LastIndexAny(adr, ":")
	if n > 0 {
		return adr[:n]
	}

	return adr
}

func (con *Connection) GetUserAgent() string {

	hdr := con.rootCon.Request().Header
	if hdr != nil {
		return hdr.Get("user-agent")
	}
	return ""
}

func (con *Connection) GetHeader(name string) string {

	hdr := con.rootCon.Request().Header
	if hdr != nil {
		return hdr.Get(name)
	}
	return ""
}
