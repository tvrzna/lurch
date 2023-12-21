package main

import (
	"sync"

	"golang.org/x/net/websocket"
)

type WsService struct {
	c     *Context
	mutex *sync.Mutex
	cons  []*WsConnection
}

type WsConnection struct {
	id  string
	con *websocket.Conn
	msg chan string
}

func NewWebSocketService(c *Context) *WsService {
	w := &WsService{c: c, mutex: &sync.Mutex{}}
	c.wsService = w
	return w
}

func (w *WsService) HandleWebSocket(con *websocket.Conn) {
	wsCon := &WsConnection{con: con, msg: make(chan string), id: randomToken(8)}

	w.mutex.Lock()
	w.cons = append(w.cons, wsCon)
	w.mutex.Unlock()

	defer con.Close()

	for {
		msg := <-wsCon.msg
		if err := websocket.Message.Send(con, msg); err != nil {
			w.removeFromSlice(wsCon)
			break
		}
	}
}

func (w *WsService) Broadcast(msg string) {
	for _, wsCon := range w.cons {
		wsCon.msg <- msg
	}
}

func (w *WsService) removeFromSlice(wsCon *WsConnection) {
	w.mutex.Lock()
	if index := w.indexOf(wsCon); index >= 0 {
		if len(w.cons) == 1 {
			w.cons = w.cons[:0]
		} else if index+1 == len(w.cons) {
			w.cons = w.cons[:index]
		} else {
			w.cons = append(w.cons[:index], w.cons[index+1])
		}
	}
	w.mutex.Unlock()
}

func (w *WsService) indexOf(wsCon *WsConnection) int {
	for i, con := range w.cons {
		if wsCon.Equals(con) {
			return i
		}
	}
	return -1
}

func (w *WsConnection) Equals(other *WsConnection) bool {
	return w.id == other.id
}
