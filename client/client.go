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

	var file, fileErr = os.OpenFile("log/"+*logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
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
                return
            }
            if replyErr != nil {
                log.Fatalf("Failed to receive message")
            }

			var oldLamport = lamport
			var newLamport = max(lamport, reply.Lamport) + 1
			lamport = newLamport

			log.Printf("[Old: %d, Server: %d, New: %d] Server: %s\n", oldLamport, reply.Lamport, newLamport, reply.Log)
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
		if strings.HasPrefix(str, "-quit") {
			var err = stream.CloseSend()
			if err != nil {
				log.Fatalf("Failed to close stream")
			}
			return
		}

		var message = &pb.Message{
			ClientName: *name,
			Message:    str,
			Lamport:    lamport,
		}

		var sendErr = stream.Send(message)
		if sendErr != nil {
			log.Panicf("Failed to send message")
		}
	}
}
