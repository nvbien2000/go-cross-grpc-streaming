# Go gRPC Bidirectional Streaming - cross communication

Recently, we encountered a problem that an action **cannot** be completed under server-side, but it needs to be done from client-side. After we tried and failed many proposals, we come to this solution: using gRPC bidirectional streaming to achieve it. Here is simple diagram:

# First

<img src="https://github.com/nvbien2000/go-cross-grpc-streaming/raw/main/assets/unary.png" />

---

# Fix

<img src="https://github.com/nvbien2000/go-cross-grpc-streaming/raw/main/assets/cross-communication.png" />

---

# Result

<img src="https://github.com/nvbien2000/go-cross-grpc-streaming/raw/main/assets/result.png" />
