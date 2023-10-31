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
	port         string

	lamport      int32
	lamportMutex sync.Mutex

    // Go maps are not thread safe
    connections  map[string]chan *pb.ChatLog
    connectionsMutex sync.Mutex
}

func (server *Server) updateLamport(lamport int32) (int32, int32) {
    server.lamportMutex.Lock()
    var oldLamport = server.lamport
    var newLamport = max(server.lamport, lamport) + 1
    server.lamport = newLamport
    server.lamportMutex.Unlock()
    return oldLamport, newLamport
}

func (server *Server) insertChannel(clientName string, channel chan *pb.ChatLog) {
    server.connectionsMutex.Lock()
    server.connections[clientName] = channel
    server.connectionsMutex.Unlock()
}

func (server *Server) sendToChannels(clientName string, chatLog *pb.ChatLog) {
    server.connectionsMutex.Lock()
    for name, channel := range server.connections {
        if name == clientName {
            continue
        }

        select {
        case channel <- chatLog:
        default:
        }
    }
    server.connectionsMutex.Lock()
}

func (server *Server) deleteChannel(clientName string) {
    server.connectionsMutex.Lock()
    var channel = server.connections[clientName]
    if channel != nil {
        close(channel)
    }
    delete(server.connections, clientName)
    server.connectionsMutex.Unlock()
}

type NoNameError struct{}
func (e *NoNameError) Error() string {
    return fmt.Sprintf("Client name is empty")
}

func (server *Server) SendChatMessages(stream pb.ChittyChat_SendChatMessagesServer) error {
    var clientName string = ""

	for {
		var msg, msgErr = stream.Recv()
		if msgErr == io.EOF {
            server.deleteChannel(clientName)
            fmt.Printf("%s has left the chat\n", msg.ClientName)
            return nil
		}
		if msgErr != nil {
			return msgErr
		}

        var oldLamport, newLamport = server.updateLamport(msg.Lamport)

		log.Printf("[Old: %d, Client: %d, New: %d] %s: %s\n", oldLamport, msg.Lamport, newLamport, msg.ClientName, msg.Message)
        
        if clientName == "" {

            if msg.ClientName == "" {
                log.Printf("Client name is empty")
                return &NoNameError{}
            }
            clientName = msg.ClientName

            var channel = make(chan *pb.ChatLog)
            server.insertChannel(clientName, channel)

            go func() {
                for {
                    var chatLog = <- channel
                    stream.Send(chatLog)
                }
            }()

            fmt.Printf("%s has joined the chat\n", msg.ClientName)

        } else { 

            var chatMsg = fmt.Sprintf("<%s>: %s", msg.ClientName, msg.Message)
            fmt.Println(chatMsg)
            var chatLog = &pb.ChatLog{
                Log: chatMsg,
                Lamport: newLamport,
            }

            server.sendToChannels(clientName, chatLog)
        }
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
            connections: make(map[string]chan *pb.ChatLog),
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
