package main

import (
	"fmt"
	"log"
	"net/http"
	"sse/bootstrap"
	apiGprc "sse/internal/api/grpc"
	apiHttp "sse/internal/api/http"
	"sse/pkg/config"
	"sse/pkg/heartbeat"
	"strconv"
	"sync"
)

// TIP <p>To run your code, right-click the code and select <b>Run</b>.</p> <p>Alternatively, click
// the <icon src="AllIcons.Actions.Execute"/> icon in the gutter and select the <b>Run</b> menu item from here.</p>
var server *http.Server

func main() {
	cfg := config.Config

	container := bootstrap.NewContainer()
	apiHttp.RegisterRoutes(container.ShardedHub)
	newHeartbeat := heartbeat.NewHeartbeat(cfg.Sse.HeartbeatSec, container.ShardedHub)
	newHeartbeat.Start()

	// 使用WaitGroup来同步gRPC和HTTP服务器的启动
	var wg sync.WaitGroup
	if cfg.Grpc.Enabled {
		apiGprc.Run(&wg, container.ShardedHub)
		wg.Add(1)
	}

	wg.Add(1)

	// 启动HTTP服务器的goroutine
	go func() {
		defer wg.Done()
		server = &http.Server{
			Addr:         ":" + strconv.Itoa(cfg.Server.Addr), // 替换为您的HTTP服务器监听的端口
			ReadTimeout:  0,
			WriteTimeout: 0,
		}
		if err := server.ListenAndServe(); err != nil {
			log.Fatalf("failed to serve HTTP: %s", err)
		}
	}()

	// 等待两个服务器完成
	wg.Wait()
	fmt.Println("Servers are running...")
}
