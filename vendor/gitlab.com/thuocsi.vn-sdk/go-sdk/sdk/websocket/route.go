package websocket

import (
	"errors"
	"sync"
)

type OnWSConnectedHandler = func(conn *Connection)
type OnWSMessageHandler = func(conn *Connection, message string)
type OnWSCloseHandler = func(conn *Connection, err error)

type wsRoute struct {
	OnConnected OnWSConnectedHandler
	OnMessage   OnWSMessageHandler
	OnClose     OnWSCloseHandler

	conMapMutex sync.Mutex
	conMap      map[int]*Connection
	payloadSize int
}

// default construction
func newWSRoute() *wsRoute {
	return &wsRoute{
		conMap:      map[int]*Connection{},
		payloadSize: 1024, // default 1024
		conMapMutex: sync.Mutex{},
	}
}

// transferring keep chatting with client
func (wsr *wsRoute) transferring(con *Connection) {

	defer func(){
		if rec := recover(); rec != nil {
			if con != nil {
				con.Deactive()
				wsr.removeCon(con.Id)
				if wsr.OnClose != nil {
					wsr.OnClose(con, errors.New("panic & recovered"))
				}
			}
		}
	}()

	for {
		if con.rootCon.MaxPayloadBytes != wsr.payloadSize {
			con.rootCon.MaxPayloadBytes = wsr.payloadSize
		}
		payload, err := con.Read()
		if err != nil {
			con.Deactive()
			wsr.removeCon(con.Id)
			if wsr.OnClose != nil {
				wsr.OnClose(con, err)
			}
			return
		} else {
			if wsr.OnMessage != nil {
				wsr.OnMessage(con, payload)
			}
		}
	}
}

func (wsr *wsRoute) removeCon(id int) {
	wsr.conMapMutex.Lock()
	delete(wsr.conMap, id)
	wsr.conMapMutex.Unlock()
}

func (wsr *wsRoute) addCon(con *Connection) {
	wsr.conMapMutex.Lock()
	wsr.conMap[con.Id] = con
	wsr.conMapMutex.Unlock()
}

func (wsr *wsRoute) GetConnectionMap() map[int]*Connection {
	wsr.conMapMutex.Lock()
	var tempMap = make(map[int]*Connection, len(wsr.conMap))
	for id, conn := range wsr.conMap {
		tempMap[id] = conn
	}
	wsr.conMapMutex.Unlock()
	return tempMap
}

func (wsr *wsRoute) GetConnection(id int) *Connection {
	return wsr.conMap[id]
}

func (wsr *wsRoute) SetPayloadSize(size int) {
	wsr.payloadSize = size

}
