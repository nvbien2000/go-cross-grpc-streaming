# Two-Way Communication Flow

## Mô tả
Luồng giao tiếp hai chiều giữa `Start()` gRPC và `PingPong()` gRPC:

1. **Start() gRPC** gửi `proto.PingRequest` cho **PingPong() gRPC**
2. **PingPong() gRPC** nhận `proto.PingRequest` rồi stream về **Client**
3. Sau khi **Client** phản hồi, **PingPong() gRPC** gửi `proto.PongResp` đó ngược lại về **Start() gRPC**

## Kiến trúc

```
┌─────────────┐    StartToPingPong    ┌─────────────┐    Stream    ┌─────────────┐
│   Start()   │ ─────────────────────► │  PingPong() │ ───────────► │   Client    │
│   gRPC      │                        │    gRPC     │              │             │
└─────────────┘                        └─────────────┘              └─────────────┘
       ▲                                      │                            │
       │                                      │                            │
       │                              PingPongToStart                      │
       │                                      │                            │
       └──────────────────────────────────────┘                            │
                                                                           │
                                                                           ▼
                                                                   PongResp
```

## Channels được sử dụng

### 1. `StartToPingPong` channel
- **Kiểu**: `chan *proto.PingReq`
- **Mục đích**: Gửi `PingReq` từ `Start()` đến `PingPong()`
- **Buffer**: 1 message

### 2. `PingPongToStart` channel  
- **Kiểu**: `chan *proto.PongResp`
- **Mục đích**: Gửi `PongResp` từ `PingPong()` về `Start()`
- **Buffer**: 1 message

## Luồng hoạt động chi tiết

### Bước 1: Start() gRPC được gọi
```go
// Start() tạo PingReq và gửi qua channel
pingReq := &proto.PingReq{
    Message: fmt.Sprintf("Ping from Start() at %s", time.Now().Format(time.RFC3339)),
}
s.StartToPingPong <- pingReq
```

### Bước 2: PingPong() nhận PingReq
```go
case pingReq := <-s.StartToPingPong:
    // Stream PingReq đến client
    stream.Send(pingReq)
```

### Bước 3: Client phản hồi
```go
// Client nhận PingReq và gửi PongResp
pongResp := &proto.PongResp{
    Message: fmt.Sprintf("Pong từ client lúc %s", time.Now().Format(time.RFC3339)),
}
stream.Send(pongResp)
```

### Bước 4: PingPong() nhận PongResp từ client
```go
case pongResp := <-clientMsgChan:
    // Gửi PongResp về Start()
    s.PingPongToStart <- pongResp
```

### Bước 5: Start() nhận PongResp
```go
case pongResp := <-s.PingPongToStart:
    log.Printf("Received PongResp from PingPong(): %s", pongResp.Message)
```

## Cách chạy test

### Terminal 1: Chạy server
```bash
make server-run
```

### Terminal 2: Chạy test flow
```bash
make test-flow
```

## Logs mong đợi

### Server logs:
```
Start() RPC called
Sending PingReq to PingPong(): Ping from Start() at 2024-01-01T12:00:00Z
PingReq sent to PingPong() successfully
Waiting for PongResp from PingPong()...
Received PingReq from Start(): Ping from Start() at 2024-01-01T12:00:00Z
Streaming PingReq to client...
PingReq sent to client successfully
Waiting for PongResp from client...
Received PongResp from client: Pong từ client lúc 2024-01-01T12:00:01Z
Sending PongResp back to Start()...
PongResp sent to Start() successfully
Received PongResp from PingPong(): Pong từ client lúc 2024-01-01T12:00:01Z
Start() RPC completed
```

### Client logs:
```
📡 Tạo PingPong stream...
🚀 Gọi Start() RPC để kích hoạt luồng...
📨 Client nhận PingReq: Ping from Start() at 2024-01-01T12:00:00Z
📤 Client gửi PongResp: Pong từ client lúc 2024-01-01T12:00:01Z
✅ Start() RPC hoàn thành thành công!
🎉 Test hoàn thành!
```

## Lợi ích của thiết kế này

1. **Giao tiếp hai chiều**: `Start()` có thể nhận phản hồi từ `PingPong()`
2. **Thread-safe**: Sử dụng channels để đồng bộ hóa
3. **Timeout handling**: Có timeout để tránh deadlock
4. **Error handling**: Xử lý lỗi một cách graceful
5. **Backward compatibility**: Vẫn giữ lại luồng cũ (legacy) 