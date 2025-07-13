package main

import (
	"context"
	"fmt"
	"go-cross-grpc-streaming/src/proto"
	"io"
	"log"
	"time"

	"google.golang.org/grpc"
)

func main() {
	// dial gRPC server
	conn, err := grpc.Dial(":50006", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("can not connect with server %v", err)
	}

	// create grpc client
	client := proto.NewPingPongClient(conn)

	// create PingPong stream FIRST
	stream, err := client.PingPong(context.Background())
	if err != nil {
		log.Fatalf("open stream error %v", err)
	}

	// call Start() rpc IMMEDIATELY to trigger PingPong stream
	log.Println("Calling Start() RPC to trigger PingPong stream...")
	_, err = client.Start(context.Background(), &proto.Empty{})
	if err != nil {
		log.Fatalf("failed to call Start rpc: %v", err)
	}
	log.Println("Start() RPC called successfully")

	ctx := stream.Context()
	done := make(chan bool)
	go func() {
		for {
			// close the stream when context is done
			req, err := stream.Recv()
			if err == io.EOF {
				log.Println("EOF received")
				close(done)
				return
			}
			if err != nil {
				log.Fatalf("can not receive %v", err)
			}

			// receive message from server
			log.Printf("new message '%s' received from server", req.Message)

			// echo back the message
			time.Sleep(1 * time.Second)
			resp := proto.PongResp{Message: fmt.Sprintf("pong %s", time.Now().Format(time.RFC3339))}
			err = stream.Send(&resp)
			if err != nil {
				log.Fatalf("can not send %v", err)
			}
		}
	}()

	// this goroutine closes done channel if context is done
	go func() {
		<-ctx.Done()
		if err := ctx.Err(); err != nil {
			log.Println(err)
		}
		close(done)
	}()

	<-done
}
