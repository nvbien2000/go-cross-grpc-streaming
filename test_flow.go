package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"go-cross-grpc-streaming/src/proto"

	"google.golang.org/grpc"
)

func main() {
	fmt.Println("=== Testing Two-Way Communication Flow ===")
	fmt.Println("1. Start() gRPC gửi proto.PingRequest cho PingPong() gRPC")
	fmt.Println("2. PingPong() nhận proto.PingRequest rồi stream về Client")
	fmt.Println("3. Sau khi client phản hồi, PingPong() gửi proto.PongResp đó ngược lại về gRPC Start()")
	fmt.Println()

	// Kết nối đến server
	conn, err := grpc.Dial(":50006", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Không thể kết nối với server: %v", err)
	}
	defer conn.Close()

	client := proto.NewPingPongClient(conn)

	// Tạo PingPong stream trước
	fmt.Println("📡 Tạo PingPong stream...")
	stream, err := client.PingPong(context.Background())
	if err != nil {
		log.Fatalf("Lỗi tạo stream: %v", err)
	}

	// Goroutine để nhận messages từ server và phản hồi
	go func() {
		for {
			pingReq, err := stream.Recv()
			if err != nil {
				log.Printf("Lỗi nhận message: %v", err)
				return
			}

			fmt.Printf("📨 Client nhận PingReq: %s\n", pingReq.Message)

			// Phản hồi lại server
			time.Sleep(1 * time.Second)
			pongResp := &proto.PongResp{
				Message: fmt.Sprintf("Pong từ client lúc %s", time.Now().Format(time.RFC3339)),
			}

			fmt.Printf("📤 Client gửi PongResp: %s\n", pongResp.Message)
			if err := stream.Send(pongResp); err != nil {
				log.Printf("Lỗi gửi PongResp: %v", err)
				return
			}
		}
	}()

	// Đợi một chút để stream được thiết lập
	time.Sleep(2 * time.Second)

	// Gọi Start() RPC để kích hoạt luồng
	fmt.Println("\n🚀 Gọi Start() RPC để kích hoạt luồng...")
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	_, err = client.Start(ctx, &proto.Empty{})
	if err != nil {
		log.Fatalf("Lỗi gọi Start RPC: %v", err)
	}

	fmt.Println("✅ Start() RPC hoàn thành thành công!")

	// Đợi một chút để xem kết quả
	time.Sleep(3 * time.Second)
	fmt.Println("\n🎉 Test hoàn thành!")
}
