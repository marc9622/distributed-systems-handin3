package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"sync"

	pb "github.com/marc9622/distributed-systems-handin3/proto"
	"google.golang.org/grpc"
)

type Server struct {
    pb.UnimplementedChittyChatServer
    port string
    lamport int32
    mutex sync.Mutex
}

func (server *Server) SendChatMessage(ctx context.Context, greeting *pb.Message) (*pb.Response, error) {

    server.mutex.Lock()
    var oldLamport = server.lamport
    var newLamport = max(server.lamport, greeting.Lamport) + 1
    server.lamport = newLamport
    server.mutex.Unlock()

    fmt.Printf("[Old: %d, Client: %d, New: %d] Received: %v\n", oldLamport, greeting.Lamport, newLamport, greeting.Message)

    return &pb.Response{
        Message: "Bye bye",
        Lamport: server.lamport,
    }, nil
}

func main() {
    var port = flag.String("port", "8080", "The port to listen on")
    flag.Parse()

    fmt.Println("Starting Server...")

    var listener, err = net.Listen("tcp", fmt.Sprintf("localhost:%s", *port))
    if err != nil {
        fmt.Printf("Failed to listen: %v", err)
        return
    }

    /* Settings up gRPC server */ {
        var grpcServer = grpc.NewServer()

        server := &Server{
            port: *port,
            lamport: 0,
        }

        pb.RegisterChittyChatServer(grpcServer, server)
        fmt.Printf("Listening on port: %s\n", *port)

        var err = grpcServer.Serve(listener)
        if err != nil {
            fmt.Printf("Failed to serve: %s", err)
            grpcServer.Stop()
            return
        }
    }

    return
}

