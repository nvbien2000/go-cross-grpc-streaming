package main

import (
	"context"
	"log"
	"net"
	"sync"
	"time"

	"go-cross-grpc-streaming/src/proto"

	"google.golang.org/grpc"
)

// Server is the server for the PingPong service.
type Server struct {
	proto.UnimplementedPingPongServer
	Mu sync.Mutex
	// PingPongStream is the single stream
	PingPongStream proto.PingPong_PingPongServer
	// PingReqChan is used to send ping request from Start() --> PingPong() --> client
	PingReqChan chan *proto.PingReq
	// PongRespChan is used to send pong response from client -- PingPong() --> Start()
	PongRespChan chan *proto.PongResp
}

// NewServer creates new instance of Server
func NewServer() *Server {
	return &Server{
		PingReqChan:  make(chan *proto.PingReq),
		PongRespChan: make(chan *proto.PongResp, 10), // Buffer size 10
	}
}

// Start is the implementation of Start() RPC
func (s *Server) Start(ctx context.Context, req *proto.Empty) (*proto.Empty, error) {
	// Trigger PingPong stream via channels
	s.Mu.Lock()
	defer s.Mu.Unlock()

	// Send ping from server to client via PingPong stream
	hasStream := s.PingPongStream != nil
	if !hasStream {
		log.Printf("No active PingPong stream")
	} else {
		log.Printf("Triggering PingPong stream via channel from Start() RPC")
		s.PingReqChan <- &proto.PingReq{Message: "ping from Start to PingPong to client"}

		// Wait for response from client (with timeout)
		select {
		case pongResp := <-s.PongRespChan:
			log.Printf("Start() received pongResp from client: %v", pongResp)
		case <-time.After(5 * time.Second):
			log.Printf("Start() timeout waiting for client response")
		}
	}

	log.Println("Start() RPC completed")
	return &proto.Empty{}, nil
}

// PingPong is the implementation of PingPong() RPC
func (s *Server) PingPong(stream proto.PingPong_PingPongServer) error {
	log.Println("started PingPong stream")

	// Register this stream (will replace any existing stream)
	s.Mu.Lock()
	s.PingPongStream = stream
	s.Mu.Unlock()

	// Cleanup when stream ends
	defer func() {
		s.Mu.Lock()
		s.PingPongStream = nil
		s.Mu.Unlock()
		log.Println("PingPong stream ended and cleaned up")
	}()

	// Channel to handle client messages
	clientMsgChan := make(chan *proto.PongResp, 1)
	clientErrChan := make(chan error, 1)

	// Goroutine to handle client messages
	go func() {
		for {
			pongResp, err := stream.Recv()
			if err != nil {
				clientErrChan <- err
				return
			}
			clientMsgChan <- pongResp
		}
	}()

	ctx := stream.Context()

	for {
		select {
		// Context done
		case <-ctx.Done():
			return ctx.Err()
		// PingPong stream triggered by Start method
		case pingReq := <-s.PingReqChan:
			log.Println("PingPong stream triggered by Start method")
			stream.Send(pingReq)
		// Received message from client
		case pongResp := <-clientMsgChan:
			log.Printf("received new pongResp=%v from client", pongResp)
			// Forward to Start() method if it's waiting
			select {
			case s.PongRespChan <- pongResp:
				log.Printf("Forwarded pongResp to Start() method")
			default:
				log.Printf("No Start() method waiting for response, pongResp dropped")
			}
		// Error from client
		case err := <-clientErrChan:
			if err.Error() == "EOF" {
				log.Println("Client disconnected")
				return nil
			}
			log.Printf("receive error %v", err)
			return err
		}
	}
}

func main() {
	// create listener
	lis, err := net.Listen("tcp", ":50006")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// create grpc server
	s := grpc.NewServer()
	proto.RegisterPingPongServer(s, NewServer())

	// start grpc server
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
