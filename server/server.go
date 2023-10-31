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
    connections  map[string]chan string
    connectionsMutex sync.Mutex
}

func (server *Server) getLamportSync() int32 {
    server.lamportMutex.Lock()
    var lamport = server.lamport
    server.lamportMutex.Unlock()
    return lamport
}

func (server *Server) updateLamportSync(lamport int32) (int32, int32) {
    server.lamportMutex.Lock()
    var oldLamport = server.lamport
    var newLamport = max(server.lamport, lamport) + 1
    server.lamport = newLamport
    server.lamportMutex.Unlock()
    return oldLamport, newLamport
}

func (server *Server) insertChannelSync(clientName string, channel chan string) {
    server.connectionsMutex.Lock()
    server.connections[clientName] = channel
    server.connectionsMutex.Unlock()
}

func (server *Server) sendToChannelsSync(clientName string, log string) {
    server.connectionsMutex.Lock()
    for _, channel := range server.connections {
        //if name == clientName {
        //    continue
        //}
        select {
        case channel <- log:
        default:
        }
    }
    server.connectionsMutex.Unlock()
}

func (server *Server) deleteChannelSync(clientName string) {
    server.connectionsMutex.Lock()
    var channel = server.connections[clientName]
    if channel != nil {
        close(channel)
    }
    delete(server.connections, clientName)
    server.connectionsMutex.Unlock()
}

func (server *Server) clientJoin(clientName string, clientLamport int32) (chan string, chan struct{}) {
    var channel = make(chan string)
    var closed = make(chan struct{})
    server.insertChannelSync(clientName, channel)
    var message = fmt.Sprintf("%s has joined the chat", clientName)
    //fmt.Println(message)
    server.sendMessageExcept(clientName, clientLamport, message)
    return channel, closed
}

func (server *Server) clientLeave(clientName string, clientLamport int32, closed chan struct{}) {
    // Close thread listening on server.connections
    select {
    case closed <- struct{}{}:
    default:
    }
    server.deleteChannelSync(clientName)
    var message = fmt.Sprintf("%s has left the chat", clientName)
    //fmt.Println(message)
    server.sendMessageExcept(clientName, clientLamport, message)
}

func (server *Server) clientChatted(clientName string, clientLamport int32, message string) {
    var msg = fmt.Sprintf("<%s>: %s", clientName, message)
    //fmt.Println(msg)
    server.sendMessageExcept(clientName, clientLamport, msg)
}

func (server *Server) logMessage(oldLamport, clientLamport, newLamport int32, clientName, message string) {
    log.Printf("[Old: %d, Client: %d, New: %d] %s: %s", oldLamport, clientLamport, newLamport, clientName, message)
}

func (server *Server) sendMessageExcept(clientName string, clientLamport int32, message string) {
    var oldLamport, newLamport = server.updateLamportSync(clientLamport)
    server.logMessage(oldLamport, clientLamport, newLamport, clientName, message)
    server.sendToChannelsSync(clientName, message)
}

type NoNameError struct{}
func (e *NoNameError) Error() string {
    return fmt.Sprintf("Client name is empty")
}

func (server *Server) SendChatMessages(stream pb.ChittyChat_SendChatMessagesServer) error {
    var clientName string = ""
    var channel chan string = nil
    //var closed chan struct{} = nil

	for {
		var msg, msgErr = stream.Recv()
		if msgErr == io.EOF {
            server.clientLeave(clientName, -1, nil)//closed)
            return nil
		}
		if msgErr != nil {
			return msgErr
		}

        if msg.IsCommand {
            switch msg.Message {
            case "-": // Client is still alive
            case "-quit":
                server.clientLeave(clientName, msg.Lamport, nil)//closed)
                return nil
            default:
                log.Printf("Unknown command: %s\n", msg.Message)
            }

            continue;
        }

        if clientName == "" {
            if msg.ClientName == "" {
                log.Printf("Client name is empty")
                return &NoNameError{}
            }
            clientName = msg.ClientName

            channel, _/*closed*/ = server.clientJoin(clientName, msg.Lamport)

            go func() {
                for {
                    select {
                    case log := <- channel:
                        var chatLog = &pb.ChatLog{
                            Log: log,
                            Lamport: server.getLamportSync(),
                        }
                        var sendErr = stream.Send(chatLog)
                        if sendErr != nil {
                            return
                        }
                    //case <- closed: // Will receive when client leaves
                    //    return
                    }
                }
            }()

            continue
        }

        server.clientChatted(clientName, msg.Lamport, msg.Message)
	}
}

func main() {
	var port = flag.String("port", "8080", "The port to listen on")
	var logFile = flag.String("log", "server.log", "The log file of the server")
	flag.Parse()

	var file, fileErr = os.OpenFile("log/"+*logFile, os.O_CREATE|os.O_WRONLY|os.O_CREATE|os.O_SYNC, 0666)
	if fileErr != nil {
		log.Panicf("Failed to open log file: %s", fileErr)
	}
	defer file.Close()

	//log.SetOutput(file)

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
            connections: make(map[string]chan string),
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
