package grpc

import (
	"net"
	"time"

	"github.com/ssimpl/simple-storage/pkg/storage"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const connectionTimeout = time.Second * 10

type Server struct {
	addr   string
	server *grpc.Server
}

func NewServer(addr string, storageSrv storage.StorageServer) *Server {
	server := grpc.NewServer(
		grpc.ConnectionTimeout(connectionTimeout),
	)

	storage.RegisterStorageServer(server, storageSrv)
	reflection.Register(server)

	return &Server{
		addr:   addr,
		server: server,
	}
}

func (s *Server) Start() error {
	listener, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}
	return s.server.Serve(listener)
}

func (s *Server) Stop() {
	s.server.GracefulStop()
}
