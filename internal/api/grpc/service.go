package __

import (
	"google.golang.org/grpc"
	"log"
	"net"
	"sse/internal/ports"
	"sync"
)

func Run(wg *sync.WaitGroup, hub ports.Hub) {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen on port 50051: %v", err)
	}
	grpcServer := grpc.NewServer()
	RegisterMessageServiceServer(grpcServer, &Server{Hub: hub})
	// 启动gRPC服务器的goroutine
	go func() {
		defer wg.Done()
		log.Println("gRPC server is running on :50051")
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve gRPC: %s", err)
		}
	}()
}
