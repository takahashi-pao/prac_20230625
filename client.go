package main

import (
	"github.com/gorilla/websocket"
)

// チャットを行っている1人のユーザーを表す。
type client struct {

	// このクライアントのためのWebSocket
	socket *websocket.Conn

	// メッセージが送られるチャネル
	send chan []byte

	// このクライアントが参加しているチャットルーム
	room *room
}

func (c *client) read() {
	defer c.socket.Close()
	for {
		_, msg, err := c.socket.ReadMessage()
		if err != nil {
			return
		}
		c.room.forward <- msg
	}
}

func (c *client) write() {
	defer c.socket.Close()
	for msg := range c.send {
		err := c.socket.WriteMessage(websocket.TextMessage, msg)
		if err != nil {
			return
		}
	}
}
