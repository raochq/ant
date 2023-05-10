package game

import (
	"context"
	"fmt"
	"testing"
	"time"

	"google.golang.org/grpc"

	"github.com/raochq/ant/protocol/pb"
)

func TestRPCServer_Echo(t *testing.T) {
	var cli pb.GameServiceClient
	// var kacp = keepalive.ClientParameters{
	// 	Time:                10 * time.Second, // send pings every 10 seconds if there is no activity
	// 	Timeout:             time.Second,      // wait 1 second for ping ack before considering the connection dead
	// 	PermitWithoutStream: true,             // send pings even without active streams
	// }
	for i := 0; i < 3; i++ {
		con, err := grpc.Dial("127.0.0.1:4100", grpc.WithInsecure()) //grpc.WithKeepaliveParams(kacp),
		if err != nil {
			fmt.Printf("rpc client connect %d failed %v\n", i, err)
			time.Sleep(time.Second)
			continue
		}
		cli = pb.NewGameServiceClient(con)
		break
	}
	if cli == nil {
		t.Errorf("connection failed")
		return
	}
	for i := 0; i < 5; i++ {
		t.Run("testRPC", func(t *testing.T) {
			resp, err := cli.Echo(context.Background(), &pb.RPCString{
				Msg: fmt.Sprintf("Echo %d", i),
			})
			if err != nil {
				t.Errorf("Echo response error %v\n", err)
			} else {
				fmt.Println("Echo response", resp.String())
			}
		})
	}
}
