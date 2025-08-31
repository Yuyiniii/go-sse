package config

import (
	"github.com/spf13/viper"
	"os"
)

var Config = loadConfig()

type config struct {
	Server struct {
		Addr int `yaml:"addr"`
	} `yaml:"server"`

	Sse struct {
		HeartbeatSec    int `yaml:"heartbeatSec"`    // 心跳时间
		ClientChanSize  int `yaml:"clientChanSize"`  // 客户端通道大小
		WriteTimeoutSec int `yaml:"writeTimeoutSec"` // 写超时时间
	} `yaml:"sse"`

	Hub struct {
		Shards         int  `yaml:"shards"`         // 分片数
		DropSlowClient bool `yaml:"dropSlowClient"` // 是否丢弃慢客户端
	} `yaml:"hub"`

	Redis struct {
		Addr   string `yaml:"addr"`   // Redis 地址
		Passwd string `yaml:"passwd"` // Redis 密码
		Db     int    `yaml:"db"`     // 数据库

		Streams struct {
			Enabled bool `yaml:"enabled"` // 是否启用
			Maxlen  int  `yaml:"maxlen"`  // 每个 topic 的保留条数
		} `yaml:"streams"`

		Pubsub struct {
			Pattern     string `yaml:"pattern"`     // 订阅模式
			PumpWorkers int    `yaml:"pumpWorkers"` // 并行解析/分发 worker 数
		} `yaml:"pubsub"`
	} `yaml:"redis"`

	Persistence struct {
		Enabled bool   `yaml:"enabled"` // 启用持久化
		Kind    string `yaml:"kind"`    // 持久化类型
		Dsn     string `yaml:"dsn"`     // 数据源名称
		Batch   struct {
			Enabled         bool `yaml:"enabled"`         // 批处理是否启用
			MaxItems        int  `yaml:"maxItems"`        // 最大条目数
			MaxBytes        int  `yaml:"maxBytes"`        // 最大字节数
			FlushIntervalMs int  `yaml:"flushIntervalMs"` // 刷新间隔
		} `yaml:"batch"`

		Topics struct {
			Include []string `yaml:"include"` // 包含的主题
			Exclude []string `yaml:"exclude"` // 排除的主题
		} `yaml:"topics"`

		Retention struct {
			Days int `yaml:"days"` // 保留天数
		} `yaml:"retention"`
	} `yaml:"persistence"`

	Publish struct {
		DefaultMaxlen int `yaml:"defaultMaxlen"` // 默认最大长度
		RateLimitQps  int `yaml:"rateLimitQps"`  // 限流 QPS
	} `yaml:"publish"`
}

func loadConfig() *config {
	vip := viper.New()

	path, err := os.Getwd()
	if err != nil {
		return nil
	}

	vip.AddConfigPath(path + "/config")
	vip.SetConfigName("config")
	vip.SetConfigType("yaml")

	if err := vip.ReadInConfig(); err != nil {
		panic(err)
	}
	var cfg config
	print(vip.AllSettings())
	err = vip.Unmarshal(&cfg)
	if err != nil {
		panic(err)
	}

	return &cfg
}
