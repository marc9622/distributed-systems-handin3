package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"strings"
	"sync"

	pb "github.com/marc9622/distributed-systems-handin3/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	var port = flag.String("port", "8080", "The port of the server")
	var name = flag.String("name", "unnamed", "The name of the client")
	var logFile = flag.String("log", "client.log", "The log file of the client")
	flag.Parse()

    // Append a random number if name is unnamed
    if *name == "unnamed" {
        *name = fmt.Sprintf("%s-%d", *name, rand.Intn(1000))
    }

	var file, fileErr = os.OpenFile("log/"+*logFile, os.O_CREATE|os.O_WRONLY|os.O_CREATE|os.O_SYNC, 0666)
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
    var lamportMutex sync.Mutex

    var updateLamport = func(lamport int32) (int32, int32) {
        lamportMutex.Lock()
        var oldLamport = lamport
        var newLamport = max(lamport, lamport) + 1
        lamport = newLamport
        lamportMutex.Unlock()
        return oldLamport, newLamport
    }

	var reader = bufio.NewReader(os.Stdin)
    var closed = make(chan struct{})

	var ctx = context.Background()

	var stream, streamErr = client.SendChatMessages(ctx)
	if streamErr != nil {
		log.Panicf("Failed to send message: %s", streamErr)
	}

    // Send join message
    {
        var message = &pb.Message{
            ClientName: *name,
            Message: "",
            IsCommand: false,
            Lamport: lamport,
        }

        var sendErr = stream.Send(message)
        if sendErr != nil {
            log.Panicf("Failed to send message")
        }
    }

    // Make thread that reads chat logs from server
    go func() {
        for {
            var reply, replyErr = stream.Recv()
            if replyErr == io.EOF {
                closed <- struct{}{}
                return
            }
            if replyErr != nil {
                log.Fatalf("Failed to receive message")
                return
            }

            var oldLamport, newLamport = updateLamport(reply.Lamport)

            log.Printf("[Prev: %d, Time: %d] Server: %s\n", oldLamport, /*reply.Lamport*/ newLamport, reply.Log)
			fmt.Println(reply.Log)
        }
    }()

    // Send messages to server
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

		var str = string(buffer[0:min(len(buffer), 128)])

        updateLamport(0)

        // If message starts with '-' then it is a command
        if strings.HasPrefix(str, "-") {
            if strings.HasPrefix(str, "-quit") {
                var leaveMsg = &pb.Message{
                    ClientName: *name,
                    Message:    "-quit",
                    IsCommand:  true,
                    Lamport:    lamport,
                }

                var sendErr = stream.Send(leaveMsg)
                if sendErr != nil {
                    log.Panicf("Failed to send message")
                }

                <- closed

                //var err = stream.CloseSend()
                //if err != nil {
                //    log.Fatalf("Failed to close stream")
                //}

                return
            } else {
                var aliveMsg = &pb.Message{
                    ClientName: *name,
                    Message:    "-",
                    IsCommand:  true,
                    Lamport:    lamport,
                }

                var sendErr = stream.Send(aliveMsg)
                if sendErr != nil {
                    log.Panicf("Failed to send message")
                }
                continue
            }
        }

        // Send message
		var message = &pb.Message{
			ClientName: *name,
			Message:    str,
            IsCommand:  false,
			Lamport:    lamport,
		}

		var sendErr = stream.Send(message)
		if sendErr != nil {
			log.Panicf("Failed to send message")
		}
	}
}
