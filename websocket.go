package barrage

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"net/http"
	"sync"
)

var wsUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// 客户端读写消息
type wsMessage struct {
	messageType int
	data        []byte
}

type Hub struct {
	Connections map[*WsConnection]bool
}

// 客户端连接
type WsConnection struct {
	wsSocket *websocket.Conn //　底层websocket
	inChan   chan *wsMessage // 读队列
	outChan  chan *wsMessage // 写队列

	mutex     sync.Mutex //避免重复关闭管道
	isClosed  bool
	closeChan chan byte //关闭通知
}

func (wsConn *WsConnection) wsReadLoop(hub *Hub) {
	for {
		//读一个message
		msgType, data, err := wsConn.wsSocket.ReadMessage()
		if err != nil {
			goto error
		}
		req := &wsMessage{
			msgType,
			data,
		}
		// 放入请求队列
		select {
		case wsConn.inChan <- req:
		case <-wsConn.closeChan:
			goto closed
		}
	}
error:
	wsConn.wsClose(hub)
closed:
}

func (wsConn *WsConnection) wsWriteLoop(hub *Hub) {
	for {
		select {
		//取一个应答
		case msg := <-wsConn.outChan:
			//写给websocket
			err := wsConn.wsSocket.WriteMessage(msg.messageType, msg.data)
			if err != nil {
				goto error
			}
		case <-wsConn.closeChan:
			goto closed
		}
	}
error:
	wsConn.wsClose(hub)
closed:
}

func (wsConn *WsConnection) procLoop(hub *Hub) {
	go func() {
		for {
			msg, err := wsConn.wsRead()
			if err != nil {
				fmt.Println("read fail")
				break
			}

			barrage := jsonToObject(msg.data)
			barrage.UserId = wsConn.wsSocket.RemoteAddr().String()
			SendToMQ(obejctToJson(barrage))

			for c := range hub.Connections {
				err = c.wsWrite(msg.messageType, obejctToJson(Filter(barrage)))
				if err != nil {
					fmt.Println("write fail")
					break
				}
			}
		}
	}()
}

func (hub Hub) WsHandler(c *gin.Context) {
	wsSocket, err := wsUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	wsConn := &WsConnection{
		wsSocket:  wsSocket,
		inChan:    make(chan *wsMessage, 1000),
		outChan:   make(chan *wsMessage, 1000),
		closeChan: make(chan byte),
		isClosed:  false,
	}

	hub.Connections[wsConn] = true

	//处理器
	go wsConn.procLoop(&hub)
	//读协程
	go wsConn.wsReadLoop(&hub)
	//写协程
	go wsConn.wsWriteLoop(&hub)
}

func (wsConn *WsConnection) wsRead() (*wsMessage, error) {
	select {
	case msg := <-wsConn.inChan:
		return msg, nil
	case <-wsConn.closeChan:
	}
	return nil, fmt.Errorf("websocket closed")
}

func (wsConn *WsConnection) wsWrite(messageType int, data []byte) error {
	select {
	case wsConn.outChan <- &wsMessage{messageType, data}:
	case <-wsConn.closeChan:
		return fmt.Errorf("websocket closed")
	}
	return nil
}

func (wsConn *WsConnection) wsClose(hub *Hub) {
	delete(hub.Connections, wsConn)
	_ = wsConn.wsSocket.Close()

	wsConn.mutex.Lock()
	defer wsConn.mutex.Unlock()
	if !wsConn.isClosed {
		wsConn.isClosed = true
		close(wsConn.closeChan)
	}
}

func jsonToObject(str []byte) Barrage {
	var barrage Barrage
	err := json.Unmarshal(str, &barrage)
	CheckError(err, "Unmarshal failed")
	return barrage
}

func obejctToJson(barrage Barrage) []byte {
	str, err := json.Marshal(barrage)
	CheckError(err, "Marshal failed")
	return str
}
