package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	pb "github.com/marc9622/distributed-systems-handin3/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
    var port = flag.String("port", "8080", "The port of the server")
    var name = flag.String("name", "unnamed", "The name of the client")
    var logFile = flag.String("log", "client.log", "The log file of the client")
    flag.Parse()

    var file, fileErr = os.OpenFile("log/" + *logFile, os.O_CREATE | os.O_WRONLY | os.O_APPEND, 0666)
    if fileErr != nil {
        log.Panicf("Failed to open log file: %s", fileErr)
    }
    defer file.Close()

    log.SetOutput(file)

    log.Printf("Starting client %s...\n", *name)

    var opt = grpc.WithTransportCredentials(insecure.NewCredentials())
    var conn, connErr = grpc.Dial(fmt.Sprintf("localhost:%s", *port), opt)
    if connErr != nil {
        log.Panicf("Failed to dial server: %s", connErr)
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
                log.Panicf("Failed to read line: %s", readErr)
            }
            buffer = append(buffer, read...)
            if !isPrefix {
                break
            }
        }

        var ctx = context.Background()

        var message = &pb.Message{
            ClientName: *name,
            Message: string(buffer[0:min(len(buffer), 128)]),
            Lamport: lamport,
        }

        var response, callErr = client.SendChatMessage(ctx, message)
        if callErr != nil {
            log.Panicf("Failed to send message: %s", callErr)
        }

        var oldLamport = lamport
        var newLamport = max(lamport, response.Lamport) + 1
        lamport = newLamport

        log.Printf("[Old: %d, Server: %d, New: %d]\n", oldLamport, response.Lamport, newLamport)
    }
}

