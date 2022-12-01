// Copyright 2022 The kubegems.io Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ws

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"kubegems.io/kubegems/pkg/log"
)

var Upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type WsMessage struct {
	MessageType int
	Data        []byte
}

type WsConnection struct {
	conn    *websocket.Conn
	inChan  chan *WsMessage
	outChan chan *WsMessage
	cancel  context.CancelFunc
	lock    sync.RWMutex
	stoped  bool
	OnClose func()
}

func (wsConn *WsConnection) wsReadLoop(ctx context.Context) {
	var (
		msgType int
		data    []byte
		err     error
	)
	for {
		if msgType, data, err = wsConn.conn.ReadMessage(); err != nil {
			log.Errorf("failed to read websocket msg %v", err)
			wsConn.WsClose()
			return
		}
		xmsg := xtermMessage{}
		e := json.Unmarshal(data, &xmsg)
		if e == nil {
			if xmsg.MsgType == "close" {
				closeMsg, _ := json.Marshal(xtermMessage{
					MsgType: "input",
					Input:   "exit\r",
				})
				wsConn.inChan <- &WsMessage{MessageType: msgType, Data: closeMsg}
			}
		}
		wsConn.inChan <- &WsMessage{MessageType: msgType, Data: data}
		select {
		case <-ctx.Done():
			return
		default:
			continue
		}
	}
}

func (wsConn *WsConnection) wsWriteLoop(ctx context.Context) {
	var (
		msg *WsMessage
		err error
	)

	for {
		select {
		case msg = <-wsConn.outChan:
			if msg != nil {
				if err = wsConn.conn.WriteMessage(msg.MessageType, msg.Data); err != nil {
					log.Errorf("failed to write websocket msg %v", err)
					wsConn.WsClose()
				}
			}
		case <-ctx.Done():
			log.Infof("stop write loop")
			return
		}
	}
}

func (wsConn *WsConnection) WsWrite(messageType int, data []byte) (err error) {
	wsConn.lock.RLock()
	defer wsConn.lock.RUnlock()
	if wsConn.stoped {
		err = errors.New("can't write on closed channel")
		return
	}
	wsConn.outChan <- &WsMessage{messageType, data}
	return
}

func (wsConn *WsConnection) WsRead() (msg *WsMessage, err error) {
	wsConn.lock.RLock()
	defer wsConn.lock.RUnlock()
	if wsConn.stoped {
		err = errors.New("can't read on closed channel")
		return
	}
	msg = <-wsConn.inChan
	return
}

func (wsConn *WsConnection) WsClose() {
	if wsConn.OnClose != nil {
		wsConn.OnClose()
	}
	wsConn.lock.Lock()
	defer wsConn.lock.Unlock()
	if wsConn.stoped {
		return
	}
	wsConn.stoped = true
	wsConn.cancel()
	wsConn.conn.Close()
	close(wsConn.inChan)
	close(wsConn.outChan)
}

func InitWebsocket(resp http.ResponseWriter, req *http.Request) (*WsConnection, error) {
	conn, err := Upgrader.Upgrade(resp, req, nil)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(context.Background())
	wsConn := &WsConnection{
		conn:    conn,
		cancel:  cancel,
		lock:    sync.RWMutex{},
		inChan:  make(chan *WsMessage, 1000),
		outChan: make(chan *WsMessage, 1000),
		stoped:  false,
	}
	go wsConn.wsReadLoop(ctx)
	go wsConn.wsWriteLoop(ctx)
	return wsConn, nil
}
