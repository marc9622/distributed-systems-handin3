package main

import (
	"context"
	"flag"
	"fmt"

	pb "github.com/marc9622/distributed-systems-handin3/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
    var port = flag.String("port", "8080", "The port of the server")
    var name = flag.String("name", "unnamed", "The name of the client")
    flag.Parse()

    fmt.Printf("Starting client %s...\n", *name)

    var opt = grpc.WithTransportCredentials(insecure.NewCredentials())
    var conn, err = grpc.Dial(fmt.Sprintf("localhost:%s", *port), opt)
    if err != nil {
        fmt.Printf("Failed to dial server: %s", err)
        return
    }
    defer conn.Close()

    /* Settings up gRPC client */ {
        var client = pb.NewChittyChatClient(conn)

        for {
            var message string
            fmt.Scanln(&message)

            var ctx = context.Background()

            var response, err = client.SayHi(ctx, &pb.Greeting{Message: message})
            if err != nil {
                fmt.Printf("Failed to send message: %s", err)
                return
            }

            fmt.Printf("Response from server: %s\n", response.Message)
        }
    }
}

