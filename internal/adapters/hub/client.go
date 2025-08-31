package hub

import (
	"sync"
	"sync/atomic"
)

type client struct {
	// 连接唯一 ID（用于管理/定位/踢线）
	id     int64
	userId int64
	// 写给该连接的数据通道（内部用，外部暴露只读视图）
	ch chan []byte
	// 关闭信号（close 后读协程退出）
	done chan struct{}
	// 原子布尔，表示是否已关闭（防止重复关闭）
	closed atomic.Bool

	// 保护 topics 等内部字段
	mu sync.RWMutex
	// 该连接当前订阅的主题集合
	topics     []string
	clientType string
}

// 创建客户端
func newClient(id int64, userId int64, buf int, clientType string, topics []string) *client {
	return &client{
		id:         id,
		userId:     userId,
		ch:         make(chan []byte, buf),
		done:       make(chan struct{}),
		topics:     topics,
		clientType: clientType,
	}
}

func (c *client) close() {
	// CAS，只有当前值等于 old 时才设置为 new
	// 只允许从 false→true 转换成功一次
	if c.closed.CompareAndSwap(false, true) {
		close(c.done)
	}
}
