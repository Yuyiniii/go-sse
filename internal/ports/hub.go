package ports

import "sync"

type Client struct {
	ID     int64
	UserId int64
	Mu     sync.Mutex
	// SendCH 仅用于 SSE 写循环读取，不在外部直接写入
	SendCh chan []byte
	// Done 在客户端关闭时关闭，上层可用户退出写循环
	Done chan struct{}
}

type HubStats struct {
	ClientID int64
	UserID   int64
	Type     string
}

type Hub interface {
	// 新建一个客户端，buf为客户端写通道缓存
	NewClient(userId int64, clientType string, topics []string) *Client

	// 广播消息到某个主题
	Broadcast(topic string, payload []byte)

	// 根据userId发送消息
	PublishByUserId(userId int64, message string)
	// 根据客户端类型发送消息
	PublishByClientType(clientType string, message string)
	// 发送到指定客户端
	PublishToClient(clientType string, userId int64, message string)
	// 根据
	// 移除连接
	Remove(c *Client)

	// 基础统计
	Stats() []HubStats

	// 心跳消息
	HeaderBeat(byte []byte)
}
