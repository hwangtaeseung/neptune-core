package grpcwrapper

import (
	"google.golang.org/grpc"
	"log"
	"net"
)

type GrpcServer struct {
	listener net.Listener
	server *grpc.Server
}

func NewGrpcServer(uri string, setupCallback func(server *grpc.Server)) *GrpcServer {
	server := &GrpcServer{}
	server.setup(uri, setupCallback)
	return server
}

func (s *GrpcServer) setup(uri string, setupCallback func(server *grpc.Server)) {
	listener, err := net.Listen("tcp", uri)
	if err != nil {
		log.Panicf("cmd listen error. %+v", err)
	}
	s.listener = listener
	s.server = grpc.NewServer()
	setupCallback(s.server)

	log.Printf("grpc cmd setup completed. (uri=%s)", uri)
}

func (s *GrpcServer) Run() {
	log.Printf("grpc cmd started. (cmd=%+v)", s)
	if err := s.server.Serve(s.listener); err != nil {
		log.Panicf("cmd run failed. %+v", err)
	}
}

func (s *GrpcServer) Shutdown() {
	s.server.Stop()
	log.Println("grpcwrapper cmd has been stopped.")
}