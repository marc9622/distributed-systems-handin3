package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
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

func (server *Server) SendChatMessage(ctx context.Context, message *pb.Message) (*pb.Response, error) {

    server.mutex.Lock()
    var oldLamport = server.lamport
    var newLamport = max(server.lamport, message.Lamport) + 1
    server.lamport = newLamport
    server.mutex.Unlock()

    log.Printf("[Old: %d, Client: %d, New: %d] %s: %v\n", oldLamport, message.Lamport, newLamport, message.ClientName, message.Message)

    return &pb.Response{
        Message: "Bye bye",
        Lamport: server.lamport,
    }, nil
}

func main() {
    var port = flag.String("port", "8080", "The port to listen on")
    var logFile = flag.String("log", "server.log", "The log file of the server")
    flag.Parse()

    var file, fileErr = os.OpenFile("log/" + *logFile, os.O_CREATE | os.O_WRONLY | os.O_APPEND, 0666)
    if fileErr != nil {
        log.Panicf("Failed to open log file: %s", fileErr)
    }
    defer file.Close()

    log.Println("Starting Server...")

    var listener, err = net.Listen("tcp", fmt.Sprintf("localhost:%s", *port))
    if err != nil {
        log.Panicf("Failed to listen: %v", err)
    }

    /* Settings up gRPC server */ {
        var grpcServer = grpc.NewServer()

        server := &Server{
            port: *port,
            lamport: 0,
        }

        pb.RegisterChittyChatServer(grpcServer, server)
        log.Printf("Listening on port: %s\n", *port)

        var err = grpcServer.Serve(listener)
        if err != nil {
            log.Panicf("Failed to serve: %s", err)
            grpcServer.Stop()
        }
    }

    return
}

