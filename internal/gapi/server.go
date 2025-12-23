package gapi

import (
	"fmt"
	"log"
	"net"

	"github.com/NhutHuyDev/sgbank/internal/infra/db"
	"github.com/NhutHuyDev/sgbank/internal/token"
	"github.com/NhutHuyDev/sgbank/pb"
	"github.com/NhutHuyDev/sgbank/pkg/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type Server struct {
	pb.UnimplementedSgbankServer
	Config     utils.Config
	Store      db.Store
	TokenMaker token.Maker
}

func NewServer(config utils.Config, store db.Store) (*Server, error) {
	tokenMaker, err := token.NewPasetoMaker(config.TokenSymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("cannot create token maker: %w", err)
	}

	server := &Server{
		Config:     config,
		Store:      store,
		TokenMaker: tokenMaker,
	}
	return server, nil
}

func (server *Server) Start(address string) error {
	grpcLogger := grpc.UnaryInterceptor(GrpcLogger)

	grpcServer := grpc.NewServer(grpcLogger)
	pb.RegisterSgbankServer(grpcServer, server)
	reflection.Register(grpcServer)

	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatal("cannot create listener", err)
		return err
	}

	log.Printf("start gRPC server at %s", listener.Addr().String())
	err = grpcServer.Serve(listener)
	if err != nil {
		log.Fatal("cannot start gRPC server", err)
		return err
	}

	return nil
}
