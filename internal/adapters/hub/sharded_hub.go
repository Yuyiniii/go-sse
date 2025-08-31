package hub

import (
	"fmt"
	"log"
	"sse/internal/ports"
	"sse/pkg/id"
	"sync"
	"sync/atomic"
)

type ShardedHub struct {

	// 全局总连接数
	totalConns int64
	// 保护 clients 映射的互斥锁
	clientsMu sync.RWMutex
	// ID 映射到内部 client 的集合，用于从外部句柄获取内部数据
	clients map[int64]*client
	// 客户端类型映射
	clientTyp map[string][]int64
	// 用户映射，用户 ID 到客户端 ID
	userMapping map[int64][]int64
}

func (h *ShardedHub) PublishByUserId(userId int64, message string) {
	int64s := h.userMapping[userId]

	for _, clients := range int64s {
		client := h.clients[clients]
		select {
		case client.ch <- []byte(message): // 尝试发送消息到客户端的通道
			fmt.Println("发送成功")
		// 发送成功，您可以选择记录日志或执行其他操作
		default:
			// 如果通道已满或者客户端未准备好，则可以考虑丢弃消息或记录
			// 这里不阻塞，如果通道关闭了，则发送将失败
			// 可以记录客户端已断开的消息，比如用 log
			log.Printf("Failed to send message to userId: %s, client might be disconnected", userId)
		}
	}

}

func (h *ShardedHub) PublishByClientType(clientType string, message string) {
	int64s := h.clientTyp[clientType]

	for _, clients := range int64s {
		client := h.clients[clients]
		select {
		case client.ch <- []byte(message): // 尝试发送消息到客户端的通道
			fmt.Println("发送成功")
			// 发送成功，您可以选择记录日志或执行其他操作
		}
	}

}

func (h *ShardedHub) PublishToClient(clientType string, userId int64, message string) {
	// 根据 userId 获取与用户相关联的客户端 ID 列表
	clientIDs := h.userMapping[userId]

	// 遍历客户端 ID，查找符合类型的客户端
	for _, clientID := range clientIDs {
		client := h.clients[clientID]

		// 检查客户端类型
		if client.clientType == clientType {
			select {
			case client.ch <- []byte(message): // 尝试发送消息到客户端的通道
				fmt.Println("Message sent successfully to client:", clientID)
			default:
				// 发送失败，记录日志
				log.Printf("Failed to send message to clientID: %d, client might be disconnected", clientID)
			}
		}
	}
}

func (h *ShardedHub) HeaderBeat(byte []byte) {
	for _, client := range h.clients {
		client.ch <- byte
	}
}

func (h *ShardedHub) Broadcast(topic string, payload []byte) {

	// 遍历所有客户端并发送消息
	for userId, client := range h.clients {
		if makeStringMap(topic, client.topics) {
			select {
			case client.ch <- []byte(payload): // 尝试发送消息到客户端的通道
				fmt.Println("发送成功")
			// 发送成功，您可以选择记录日志或执行其他操作
			default:
				// 如果通道已满或者客户端未准备好，则可以考虑丢弃消息或记录
				// 这里不阻塞，如果通道关闭了，则发送将失败
				// 可以记录客户端已断开的消息，比如用 log
				log.Printf("Failed to send message to userId: %s, client might be disconnected", userId)
			}
		}
	}
}

// ShardedHub 的 Remove 方法
func (h *ShardedHub) Remove(c *ports.Client) {
	c.Mu.Lock() // 确保安全访问
	defer c.Mu.Unlock()

	// 确保客户端存在且通道未关闭
	if c.SendCh != nil {
		close(c.SendCh) // 关闭发送通道
		c.SendCh = nil  // 设置为 nil，避免后续操作冲突
	}
	<-c.Done

	// 确保检查客户机的存在性
	if client, exists := h.clients[c.ID]; exists && client != nil {
		client.close()
		delete(h.clients, c.ID) // 从 Hub 中移除
	} else {
		log.Printf("客户端 %d 不存在或已关闭, 无法移除\n", c.ID)
	}
}

func (h *ShardedHub) Stats() []ports.HubStats {
	data := make([]ports.HubStats, 0)

	// 遍历用户映射
	for userID, clientIDs := range h.userMapping {
		for _, clientID := range clientIDs { // 遍历每一个客户端 ID
			// 检查该客户端是否在 clients 中
			if clientInfo, exists := h.clients[clientID]; exists {
				data = append(data, ports.HubStats{
					ClientID: clientID,
					UserID:   userID,
					Type:     clientInfo.clientType, // 从客户端信息获取类型
				})
			}
		}
	}

	return data
}

// 构建分片hub实例
// numShards: 分片数
func NewShardedHub() *ShardedHub {
	return &ShardedHub{
		totalConns:  0,
		clientsMu:   sync.RWMutex{},
		clients:     make(map[int64]*client),
		clientTyp:   make(map[string][]int64),
		userMapping: make(map[int64][]int64),
	}
}

// NewClient 实现 ports.Hub.NewClient，创建新的客户端并返回。
// 参数：
//   - buf: 客户端通道的缓冲区大小
//
// 返回：
//   - *ports.Client: 返回的客户端句柄供上层使用
func (h *ShardedHub) NewClient(userId int64, clientType string, topics []string) *ports.Client {
	globalID := id.NextGlobalID()
	c := newClient(globalID, userId, 255, clientType, topics) // 创建 client 实例

	atomic.AddInt64(&h.totalConns, 1) // 更新总连接数
	h.clients[globalID] = c
	h.clientTyp[clientType] = append(h.clientTyp[clientType], globalID)
	h.userMapping[userId] = append(h.userMapping[userId], globalID)

	log.Printf("添加用户:%d ,唯一ID:%d ,clientsSize:%d ,clientTyp:%d ,clientTypeSize:%d ,userMapping:%d ,userMappingSize:%d ,",
		userId, globalID, len(h.clients), clientType, len(h.clientTyp[clientType]), userId, len(h.userMapping[userId]))
	// 返回上层只读的客户端句柄
	return &ports.Client{
		ID:     globalID,
		SendCh: c.ch,
		Done:   c.done, // 返回关闭信号等待通道
	}
}

// removeValue 从切片中删除指定的值
func removeValue(slice []int64, value int64) []int64 {
	// 创建一个新的切片，用于存放不包含指定值的元素
	newSlice := make([]int64, 0)

	for _, v := range slice {
		if v != value {
			newSlice = append(newSlice, v) // 仅添加不等于 value 的元素
		}
	}

	return newSlice // 返回的新切片
}

// makeStringMap 函数接受一个字符串和一个字符串切片，返回切片中该字符串是否存在
func makeStringMap(str string, slice []string) bool {
	for _, v := range slice {
		if v == str {
			return true
		}
	}
	return false
}
