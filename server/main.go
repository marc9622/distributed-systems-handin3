package main

import (
	"context"
	"flag"
	"fmt"
	"net"

	pb "github.com/marc9622/distributed-systems-handin3/proto"
	"google.golang.org/grpc"
)

type Server struct {
    pb.UnimplementedChittyChatServer
    port string
}

func (server *Server) SayHi(ctx context.Context, greeting *pb.Greeting) (*pb.Farewell, error) {
    fmt.Printf("Received: %v\n", greeting.Message)
    return &pb.Farewell{Message: "Bye bye"}, nil
}

func main() {
    var port = flag.String("port", "8080", "The server port");
    flag.Parse();

    fmt.Println("Starting Server...");

    listener, err := net.Listen("tcp", fmt.Sprintf("localhost:%s", *port));
    if err != nil {
        fmt.Printf("Failed to listen: %v", err);
        return;
    }

    /* Settings up gRPC server */ {
        grpcServer := grpc.NewServer()

        server := &Server{
            port: *port,
        }

        pb.RegisterChittyChatServer(grpcServer, server)
        fmt.Printf("Listening on port: %s\n", *port);

        err := grpcServer.Serve(listener)
        if err != nil {
            fmt.Printf("Failed to serve: %s", err)
            return
        }
    }

}

