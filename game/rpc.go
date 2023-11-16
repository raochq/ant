package game

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"strconv"

	"github.com/raochq/ant/protocol/pb"
	"google.golang.org/grpc"
)

type RPCServer struct {
	pb.UnimplementedGameServiceServer
	owner *Game
	svr   *grpc.Server
}

func (g *RPCServer) startGrpc(rpcAddr string) (string, error) {
	lis, err := net.Listen("tcp", rpcAddr)
	if err != nil {
		return "", fmt.Errorf("GameRPCServer listen %s fail:%w", rpcAddr, err)
	}
	tcpAddr := lis.Addr().(*net.TCPAddr)
	addr := tcpAddr.String()
	if tcpAddr.IP.Equal(net.IPv4zero) || tcpAddr.IP.Equal(net.IPv6zero) {
		addr = net.JoinHostPort("localhost", strconv.Itoa(tcpAddr.Port))
	}

	//s := grpc.NewServer( grpc.UnaryInterceptor(ServiceProxy))
	g.svr = grpc.NewServer()
	pb.RegisterGameServiceServer(g.svr, g)
	go g.svr.Serve(lis)
	slog.Info("=== GameRPCServer started ===", "rpc", lis.Addr())
	return addr, nil
}

func (g *RPCServer) stopGrpc() {
	g.svr.Stop()
	slog.Info("=== GameRPCServer stopped ===")
}

func (g *RPCServer) Echo(ctx context.Context, req *pb.RPCString) (*pb.RPCString, error) {
	resp := &pb.RPCString{
		Msg: "Hello " + req.GetMsg(),
	}
	return resp, nil
}
