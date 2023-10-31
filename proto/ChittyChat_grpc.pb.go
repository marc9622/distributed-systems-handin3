// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v4.24.4
// source: proto/ChittyChat.proto

package proto

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// ChittyChatClient is the client API for ChittyChat service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ChittyChatClient interface {
	SendChatMessages(ctx context.Context, opts ...grpc.CallOption) (ChittyChat_SendChatMessagesClient, error)
}

type chittyChatClient struct {
	cc grpc.ClientConnInterface
}

func NewChittyChatClient(cc grpc.ClientConnInterface) ChittyChatClient {
	return &chittyChatClient{cc}
}

func (c *chittyChatClient) SendChatMessages(ctx context.Context, opts ...grpc.CallOption) (ChittyChat_SendChatMessagesClient, error) {
	stream, err := c.cc.NewStream(ctx, &ChittyChat_ServiceDesc.Streams[0], "/proto.ChittyChat/SendChatMessages", opts...)
	if err != nil {
		return nil, err
	}
	x := &chittyChatSendChatMessagesClient{stream}
	return x, nil
}

type ChittyChat_SendChatMessagesClient interface {
	Send(*Message) error
	CloseAndRecv() (*ChatLog, error)
	grpc.ClientStream
}

type chittyChatSendChatMessagesClient struct {
	grpc.ClientStream
}

func (x *chittyChatSendChatMessagesClient) Send(m *Message) error {
	return x.ClientStream.SendMsg(m)
}

func (x *chittyChatSendChatMessagesClient) CloseAndRecv() (*ChatLog, error) {
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	m := new(ChatLog)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// ChittyChatServer is the server API for ChittyChat service.
// All implementations must embed UnimplementedChittyChatServer
// for forward compatibility
type ChittyChatServer interface {
	SendChatMessages(ChittyChat_SendChatMessagesServer) error
	mustEmbedUnimplementedChittyChatServer()
}

// UnimplementedChittyChatServer must be embedded to have forward compatible implementations.
type UnimplementedChittyChatServer struct {
}

func (UnimplementedChittyChatServer) SendChatMessages(ChittyChat_SendChatMessagesServer) error {
	return status.Errorf(codes.Unimplemented, "method SendChatMessages not implemented")
}
func (UnimplementedChittyChatServer) mustEmbedUnimplementedChittyChatServer() {}

// UnsafeChittyChatServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ChittyChatServer will
// result in compilation errors.
type UnsafeChittyChatServer interface {
	mustEmbedUnimplementedChittyChatServer()
}

func RegisterChittyChatServer(s grpc.ServiceRegistrar, srv ChittyChatServer) {
	s.RegisterService(&ChittyChat_ServiceDesc, srv)
}

func _ChittyChat_SendChatMessages_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(ChittyChatServer).SendChatMessages(&chittyChatSendChatMessagesServer{stream})
}

type ChittyChat_SendChatMessagesServer interface {
	SendAndClose(*ChatLog) error
	Recv() (*Message, error)
	grpc.ServerStream
}

type chittyChatSendChatMessagesServer struct {
	grpc.ServerStream
}

func (x *chittyChatSendChatMessagesServer) SendAndClose(m *ChatLog) error {
	return x.ServerStream.SendMsg(m)
}

func (x *chittyChatSendChatMessagesServer) Recv() (*Message, error) {
	m := new(Message)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// ChittyChat_ServiceDesc is the grpc.ServiceDesc for ChittyChat service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var ChittyChat_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "proto.ChittyChat",
	HandlerType: (*ChittyChatServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "SendChatMessages",
			Handler:       _ChittyChat_SendChatMessages_Handler,
			ClientStreams: true,
		},
	},
	Metadata: "proto/ChittyChat.proto",
}
