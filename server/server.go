package server

import (
	"net"

	"github.com/qxnw/grpc4ars/pb"

	"golang.org/x/net/context"

	"google.golang.org/grpc"
)

//Server RPC Server
type Server struct {
	address  string
	callback func(string, string, string) (int, string, error)
}

//NewServer 初始化
func NewServer(f func(string, string, string) (int, string, error)) *Server {
	return &Server{callback: f}
}

//Start 启动RPC服务器
func (r *Server) Start(address string) (err error) {
	lis, err := net.Listen("tcp", address)
	if err != nil {
		return
	}
	s := grpc.NewServer()
	pb.RegisterARSServer(s, r)
	s.Serve(lis)
	return
}

//Request 客户端处理客户端请求
func (r *Server) Request(context context.Context, request *pb.RequestContext) (p *pb.ResponseContext, err error) {
	s, d, err := r.callback(request.Session, request.Sevice, request.Input)
	if err != nil {
		return
	}
	p = &pb.ResponseContext{Status: int32(s), Result: d}
	return
}
