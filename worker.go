package main

import (
	"log"
	"net/http"
	"net/url"

	"github.com/gorilla/websocket"
)

func main() {
	// 连接到WebSocket服务器
	u := url.URL{Scheme: "ws", Host: "localhost:8000", Path: "/worker"}
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	// 接收来自服务器的消息
	go func() {
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			// 在这里处理消息并返回结果
			result := processMessage(message)
			// 将结果发送回服务器
			err = c.WriteMessage(websocket.TextMessage, result)
			if err != nil {
				log.Println("write:", err)
				return
			}
		}
	}()

	// 处理HTTP请求
	http.HandleFunc("/", handleRequest)
	log.Fatal(http.ListenAndServe(":8081", nil))
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	// 在这里处理HTTP请求并构建响应
	// 这里只是一个示例，你需要根据实际情况实现
	response := []byte("Processed request.")
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade:", err)
		return
	}
	defer c.Close()
	c.WriteMessage(websocket.TextMessage, response)
}

func processMessage(message []byte) []byte {
	// 根据消息内容处理并返回结果
	// 这里只是一个示例，你需要根据实际情况实现
	return []byte("Processed message: " + string(message))
}
