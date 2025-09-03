package http

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sse/internal/ports"
	"strconv"
	"strings"
)

// 解析用户 ID
func parseUserID(r *http.Request) (int64, error) {
	userIdStr := r.URL.Query().Get("userId")
	userId, err := strconv.ParseInt(userIdStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("无效的 userId: %v", err)
	}
	return userId, nil
}

func parseClientType(r *http.Request) (string, error) {
	clientType := r.URL.Query().Get("clientType")
	if clientType == "" {
		return "", fmt.Errorf("clientType 查询参数不存在")
	}
	return clientType, nil
}

func parseTopics(r *http.Request) ([]string, error) {
	topicsStr := r.URL.Query().Get("topics")
	if topicsStr == "" {
		return nil, fmt.Errorf("topics 查询参数不存在")
	}

	topics := strings.Split(topicsStr, ",")
	for i := range topics {
		topics[i] = strings.TrimSpace(topics[i])
	}
	return topics, nil
}
func Sse(hub ports.Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 设置 CORS 头
		w.Header().Set("Access-Control-Allow-Origin", "*") // 允许所有来源，您也可以指定特定来源
		w.Header().Set("Access-Control-Allow-Methods", "GET")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		// 设置响应头
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		// 获取id
		userId, _ := parseUserID(r)

		clientType, err := parseClientType(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		topics, err := parseTopics(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		client := hub.NewClient(userId, clientType, topics)

		// 尝试获取 CloseNotifier
		cn, ok := w.(http.CloseNotifier)
		if !ok {
			// 如果不支持 CloseNotifier，则直接处理消息
			go handleClientMessages(w, hub, client)
			<-client.Done // 等待 Done 通道
			return
		}

		// 获取 CloseNotify() 通道
		closeNotify := cn.CloseNotify()

		// 创建 goroutine 处理消息
		go handleClientMessages(w, hub, client)

		// 监听 CloseNotifier 通道
		select {
		case <-closeNotify:
			log.Printf("客户端 %d 断开连接 (CloseNotifier)\n", client.ID)
			hub.Remove(client)
		case <-client.Done:
			log.Printf("客户端 %d 断开连接 (client.Done)\n", client.ID)
		}
	}
}

// 处理消息发送
func handleClientMessages(w http.ResponseWriter, hub ports.Hub, client *ports.Client) {
	defer func() {
		log.Printf("客户端 %d 断开连接时处理\n", client.ID)
		hub.Remove(client) // 确保在断开时移除客户端
	}()

	for msg := range client.SendCh { // 读取消息的通道
		message := fmt.Sprintf("data: %s\n\n", msg) // 格式化消息
		if _, err := w.Write([]byte(message)); err != nil {
			log.Printf("发送消息时发生错误，客户端 %d: %v\n", client.ID, err)
			break // 发送出错后退出
		}
		if f, ok := w.(http.Flusher); ok {
			f.Flush() // 确保消息立即发送到客户端
		}
	}
}
func RegisterRoutes(hub ports.Hub) {
	http.HandleFunc("/sse", Sse(hub))
	http.HandleFunc("/publishByTopic", PublishByTopic(hub))
	http.HandleFunc("/publishByUserId", PublishByUserId(hub))
	http.HandleFunc("/publishByClientType", PublishByClientType(hub))
	http.HandleFunc("/publishToClient", PublishToClient(hub))
	http.HandleFunc("/status", Status(hub))
}

type PublishToClientMessageBody struct {
	ClientType string `json:"clientType"`
	UserId     int64  `json:"userId"`
	Message    string `json:"message"`
}

func PublishToClient(hub ports.Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body PublishToClientMessageBody
		err := json.NewDecoder(r.Body).Decode(&body)
		if err != nil {
			return
		}
		hub.PublishToClient(body.ClientType, body.UserId, body.Message)
	}
}

type PublishByClientTypeMessageBody struct {
	ClientType string `json:"clientType"`
	Message    string `json:"message"`
}

func PublishByClientType(hub ports.Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body PublishByClientTypeMessageBody
		err := json.NewDecoder(r.Body).Decode(&body)
		if err != nil {
			return
		}
		hub.PublishByClientType(body.ClientType, body.Message)
	}
}

type PublishByUserIdMessageBody struct {
	UserId  int64  `json:"userId"`
	Message string `json:"message"`
}

func PublishByUserId(hub ports.Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body PublishByUserIdMessageBody
		err := json.NewDecoder(r.Body).Decode(&body)
		if err != nil {
			return
		}
		hub.PublishByUserId(body.UserId, body.Message)
	}
}

func Status(hub ports.Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 获取客户端状态
		stats := hub.Stats()

		// 设置响应内容类型为 JSON
		w.Header().Set("Content-Type", "application/json")

		// 将 stats 编码为 JSON 并写入响应
		if err := json.NewEncoder(w).Encode(stats); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

type PublishByTopicMessageBody struct {
	Topic   string `json:"topic"`
	Message string `json:"message"`
}

func PublishByTopic(hub ports.Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body PublishByTopicMessageBody
		err := json.NewDecoder(r.Body).Decode(&body)
		if err != nil {
			return
		}
		hub.Broadcast(body.Topic, []byte(body.Message))
	}
}
