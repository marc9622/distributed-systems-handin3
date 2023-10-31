package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"sync"

	pb "github.com/marc9622/distributed-systems-handin3/proto"
	"google.golang.org/grpc"
)

type Server struct {
	pb.UnimplementedChittyChatServer
	port    string
	lamport int32
	mutex   sync.Mutex
}

func (server *Server) SendChatMessages(stream pb.ChittyChat_SendChatMessagesServer) error {
	for {
		var message, err = stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&pb.ChatLog{
				Log:     "Bye bye",
				Lamport: server.lamport,
			})
		}
		if err != nil {
			return err
		}

		server.mutex.Lock()
		var oldLamport = server.lamport
		var newLamport = max(server.lamport, message.Lamport) + 1
		server.lamport = newLamport
		server.mutex.Unlock()

		log.Printf("[Old: %d, Client: %d, New: %d] %s: %s\n", oldLamport, message.Lamport, newLamport, message.ClientName, message.Message)
		fmt.Printf("<%s>: %s\n", message.ClientName, message.Message)
	}
}

func main() {
	var port = flag.String("port", "8080", "The port to listen on")
	var logFile = flag.String("log", "server.log", "The log file of the server")
	flag.Parse()

	var file, fileErr = os.OpenFile("log/"+*logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if fileErr != nil {
		log.Panicf("Failed to open log file: %s", fileErr)
	}
	defer file.Close()

	log.SetOutput(file)

	log.Println("Starting Server...")

	var listener, err = net.Listen("tcp", fmt.Sprintf("localhost:%s", *port))
	if err != nil {
		log.Panicf("Failed to listen: %v", err)
	}

	/* Settings up gRPC server */
	{
		var grpcServer = grpc.NewServer()

		server := &Server{
			port:    *port,
			lamport: 0,
		}

		pb.RegisterChittyChatServer(grpcServer, server)
		log.Printf("Listening on port: %s\n", *port)

		var err = grpcServer.Serve(listener)
		if err != nil {
			grpcServer.Stop()
			log.Panicf("Failed to serve: %s", err)
		}
	}

	return
}
