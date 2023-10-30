package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"

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
    var conn, connErr = grpc.Dial(fmt.Sprintf("localhost:%s", *port), opt)
    if connErr != nil {
        fmt.Printf("Failed to dial server: %s", connErr)
        return
    }
    defer conn.Close()

    var client = pb.NewChittyChatClient(conn)

    var lamport int32 = 0
    var reader = bufio.NewReader(os.Stdin)

    for {
        var buffer []byte
        for {
            var read, isPrefix, readErr = reader.ReadLine()
            if readErr != nil {
                fmt.Printf("Failed to read line: %s", readErr)
                return
            }
            buffer = append(buffer, read...)
            if !isPrefix {
                break
            }
        }

        var ctx = context.Background()

        var greeting = &pb.Message{
            Message: string(buffer[0:min(len(buffer), 128)]),
            Lamport: lamport,
        }

        var response, callErr = client.SendChatMessage(ctx, greeting)
        if callErr != nil {
            fmt.Printf("Failed to send message: %s", callErr)
            return
        }

        var oldLamport = lamport
        var newLamport = max(lamport, response.Lamport) + 1
        lamport = newLamport

        fmt.Printf("[Current: %d, Server: %d, New: %d] Response: %s\n", oldLamport, response.Lamport, newLamport, response.Message)
    }
}

