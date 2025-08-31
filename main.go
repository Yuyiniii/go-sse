package main

import (
	"fmt"
	"net/http"
	"sse/bootstrap"
	apiHttp "sse/internal/api/http"
	"sse/pkg/config"
	"sse/pkg/heartbeat"
	"strconv"
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
	server = &http.Server{
		Addr:         ":" + strconv.Itoa(cfg.Server.Addr),
		ReadTimeout:  0,
		WriteTimeout: 0,
	}
	err := server.ListenAndServe()
	if err != nil {
		fmt.Println(err)
		return
	}
}
