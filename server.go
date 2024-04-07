package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/ybz21/AGIOcean/models"
)

const TUNNEL_PREFIX = "/proxy"

var upgrader = websocket.Upgrader{}
var aliveWorkers = make([]models.Worker, 0)
var responseInfoChannel = make(chan models.ResponseInfo, 10000)

func webSocketHandler(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()

	worker := models.Worker{
		ID:         uuid.New().String(),
		IP:         c.ClientIP(),
		Online:     true,
		ModelName:  "default",
		Connection: conn,
	}

	aliveWorkers = append(aliveWorkers, worker)

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("Error during message reading:", err)
			//delete(aliveWorkers, worker)
			// todo: delete

			conn.Close()
			break
		}

		var responseInfo models.ResponseInfo
		err = json.Unmarshal(message, &responseInfo)
		if err != nil {
			fmt.Printf("Error during message unmarshal: %v", err)
			break
		}
		fmt.Println("responseInfo: ", responseInfo)
		responseInfoChannel <- responseInfo
	}
}

// HandleHTTPRequest receives HTTP requests and forwards them to WebSocket clients.
func tunnelHandler(c *gin.Context) {
	r := c.Request
	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println("Error reading request body:", err)
		return
	}

	requestInfo := models.RequestInfo{
		RequestID: uuid.New().String(),
		Query:     r.URL.Query(),
		Path:      r.URL.Path[len("TUNNEL_PREFIX"):],
		Header:    r.Header,
		Body:      bodyBytes,
	}

	worker, err := getWorker()
	if err != nil {
		response := map[string]interface{}{
			"error": fmt.Sprintf("Get worker error: %v", err),
		}
		c.JSON(http.StatusOK, response)
		return
	}

	err = worker.Connection.WriteJSON(requestInfo)
	if err != nil {
		log.Println("Error during message writing:", err)
	}

	// todo: 从channel中获取响应id为requestInfo.RequestID的 responseInfo
	// 创建goroutine来监听响应

	var wg sync.WaitGroup
	wg.Add(1) // 增加一个等待计数

	go func(wg *sync.WaitGroup, c *gin.Context) {
		defer wg.Done() // 确保在 goroutine 结束时减少等待计数

		var response models.ResponseInfo
		for {
			select {
			case resp := <-responseInfoChannel:
				fmt.Println("------")
				fmt.Printf("resp.RequestID: %s ,requestInfo.RequestID: %s", resp.RequestID, requestInfo.RequestID)
				if resp.RequestID == requestInfo.RequestID {
					response = resp
					genResponse(c, response)
					return
				} else {
					// 放回队列
					responseInfoChannel <- resp
				}

			case <-time.After(500 * time.Millisecond): // 避免阻塞，定期检查队列
				// 如果没有消息，休眠一段时间
				fmt.Println("timeout")
			}
		}

	}(&wg, c)

	// 等待 goroutine 完成
	wg.Wait()
}

func genResponse(c *gin.Context, responseInfo models.ResponseInfo) {
	for k, v := range responseInfo.Header {
		c.Header(k, v[0])
	}

	var body map[string]interface{}
	err := json.Unmarshal(responseInfo.Body, &body)
	if err != nil {
		response := map[string]interface{}{
			"error": fmt.Sprintf("Unmarshal response body error: %v", err),
		}
		c.JSON(responseInfo.StatusCode, response)
		return
	}
	c.JSON(responseInfo.StatusCode, body)
}

func getWorker() (models.Worker, error) {
	if len(aliveWorkers) > 0 {
		for _, worker := range aliveWorkers {
			//if worker.Ability == "default" {
			//	return worker, nil
			//}you
			fmt.Println(worker.Abilities)
		}
		return aliveWorkers[0], nil
	}

	return models.Worker{}, fmt.Errorf("no worker")
}

func listWorkersHandler(c *gin.Context) {
	c.JSON(http.StatusOK, aliveWorkers)
}

func main() {
	r := gin.Default()
	r.Any(fmt.Sprintf("%s/*path", TUNNEL_PREFIX), tunnelHandler)
	r.GET("/ws", webSocketHandler)

	r.GET("/api/v1/workers", listWorkersHandler)

	r.Run(":8080")
}
