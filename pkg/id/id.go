package id

import "sync/atomic"

// IDGenerator 用于生成唯一的 ID
type IDGenerator struct {
	id int64
}

// 实例化全局 ID 生成器
var generator = &IDGenerator{}

// NextID 返回下一个唯一的 ID
func (gen *IDGenerator) NextID() int64 {
	return atomic.AddInt64(&gen.id, 1) // 原子增加并返回新的 ID
}

// 提供一个全局方法供外部调用生成 ID
func NextGlobalID() int64 {
	return generator.NextID()
}
