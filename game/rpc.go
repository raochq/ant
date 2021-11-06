package game

import (
	"context"
	"fmt"
	"github.com/raochq/ant/engine/logger"
	"github.com/raochq/ant/protocol/pb"
	"google.golang.org/grpc"
	"net"
)

type RPCServer struct {
	pb.UnimplementedGameServiceServer
	owner *Game
	svr   *grpc.Server
}

func (g *RPCServer) startGrpc(port uint32) error {
	rpcAddr := fmt.Sprintf("0.0.0.0:%d", port)
	lis, err := net.Listen("tcp", rpcAddr)
	if err != nil {
		return fmt.Errorf("GameRPCServer listen %s fail:%w", rpcAddr, err)
	}
	//s := grpc.NewServer( grpc.UnaryInterceptor(ServiceProxy))
	g.svr = grpc.NewServer()
	pb.RegisterGameServiceServer(g.svr, g)
	go g.svr.Serve(lis)
	logger.Info("===GameRPCServer started===")
	return nil
}

func (g *RPCServer) stopGrpc() {
	g.svr.Stop()
	logger.Info("===GameRPCServer stopped===")
}

func (g *RPCServer) Echo(ctx context.Context, req *pb.RPCString) (*pb.RPCString, error) {
	resp := &pb.RPCString{
		Msg: req.GetMsg(),
	}
	return resp, nil
}
