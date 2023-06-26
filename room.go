package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type room struct {
	// 他クライアントに転送メッセージを保持するチャネル
	forward chan []byte
	// チャットルームに参加しようとしているクライアントのためのチャネル
	join chan *client
	// チャットルームから退室しようとしているクライアントのためのチャネル
	leave chan *client
	// 在室の全クライアントのためのチャネル
	clients map[*client]bool

	// tips: チャネルを使わずにマップを直接操作 → 複数のgoroutineがマップを同時に変更する可能性＝メモリの破壊、予期せぬエラーに繋がる
}

// do: ①チャットルーム内でチャネル（join, leave, forward）を監視
// do: ②いずれかのチャネルにメッセージが届く → case節が実行
func (r *room) run() {
	for {
		select {
		case client := <-r.join:
			// 参加
			r.clients[client] = true
		case client := <-r.leave:
			// 退室
			delete(r.clients, client)
			close(client.send)
		case msg := <-r.forward:
			// 全クライアントにメッセージを転送
			for client := range r.clients {
				select {
				case client.send <- msg:
					// メッセージを送信
				default:
					// 送信に失敗
					delete(r.clients, client)
					close(client.send)
				}
			}
		}
	}
}

const (
	socketBufferSize  = 1024
	messageBufferSize = 256
)

// websocket利用のため、HTTP接続をアップグレードするための型
var upgrader = &websocket.Upgrader{
	ReadBufferSize:  socketBufferSize,
	WriteBufferSize: socketBufferSize}

func (r *room) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// websocketコネクションの取得
	socket, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		log.Fatal("SearveHTTP:", err)
		return
	}

	// clientを生成し、現在のチャットルームのjoinチャネルに渡す
	client := &client{
		socket: socket,
		send:   make(chan []byte, messageBufferSize),
		room:   r,
	}

	r.join <- client

	// クライアント終了時に退室処理を行う
	defer func() { r.leave <- client }()
	go client.write()
	client.read()

}
