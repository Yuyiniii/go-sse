package __

import (
	"context"
	"sse/internal/ports"
	// 修改为生成的Go代码的包路径
)

type Server struct {
	Hub ports.Hub // 用于处理消息的Hub
}

func (s *Server) mustEmbedUnimplementedMessageServiceServer() {
	// TODO implement me
	panic("implement me")
}

// PublishByTopic 实现
func (s *Server) PublishByTopic(ctx context.Context, req *PublishByTopicRequest) (*Empty, error) {
	s.Hub.Broadcast(req.Topic, []byte(req.Message))
	return &Empty{}, nil
}

// PublishByUserId 实现
func (s *Server) PublishByUserId(ctx context.Context, req *PublishByUserIdRequest) (*Empty, error) {
	s.Hub.PublishByUserId(req.UserId, req.Message)
	return &Empty{}, nil
}

// PublishByClientType 实现
func (s *Server) PublishByClientType(ctx context.Context, req *PublishByClientTypeRequest) (*Empty, error) {
	s.Hub.PublishByClientType(req.ClientType, req.Message)
	return &Empty{}, nil
}

// PublishToClient 实现
func (s *Server) PublishToClient(ctx context.Context, req *PublishToClientRequest) (*Empty, error) {
	s.Hub.PublishToClient(req.ClientType, req.UserId, req.Message)
	return &Empty{}, nil
}

// Status 实现
func (s *Server) Status(ctx context.Context, req *StatusRequest) (*StatusResponse, error) {
	_ = s.Hub.Stats()

	// 构建响应
	response := &StatusResponse{}
	//
	// for _, stat := range stats {
	// 	response.Stats = append(response.Stats, &ClientStat{
	// 		ClientId: stat.ClientID,
	// 		Status:   stat.Type,
	// 	})
	// }

	return response, nil
}
