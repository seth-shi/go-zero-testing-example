// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v3.19.4
// source: id.proto

package id

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

const (
	Id_Get_FullMethodName = "/id.Id/Get"
)

// IdClient is the client API for Id service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type IdClient interface {
	Get(ctx context.Context, in *IdRequest, opts ...grpc.CallOption) (*IdResponse, error)
}

type idClient struct {
	cc grpc.ClientConnInterface
}

func NewIdClient(cc grpc.ClientConnInterface) IdClient {
	return &idClient{cc}
}

func (c *idClient) Get(ctx context.Context, in *IdRequest, opts ...grpc.CallOption) (*IdResponse, error) {
	out := new(IdResponse)
	err := c.cc.Invoke(ctx, Id_Get_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// IdServer is the server API for Id service.
// All implementations must embed UnimplementedIdServer
// for forward compatibility
type IdServer interface {
	Get(context.Context, *IdRequest) (*IdResponse, error)
	mustEmbedUnimplementedIdServer()
}

// UnimplementedIdServer must be embedded to have forward compatible implementations.
type UnimplementedIdServer struct {
}

func (UnimplementedIdServer) Get(context.Context, *IdRequest) (*IdResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Get not implemented")
}
func (UnimplementedIdServer) mustEmbedUnimplementedIdServer() {}

// UnsafeIdServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to IdServer will
// result in compilation errors.
type UnsafeIdServer interface {
	mustEmbedUnimplementedIdServer()
}

func RegisterIdServer(s grpc.ServiceRegistrar, srv IdServer) {
	s.RegisterService(&Id_ServiceDesc, srv)
}

func _Id_Get_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(IdRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(IdServer).Get(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Id_Get_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(IdServer).Get(ctx, req.(*IdRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Id_ServiceDesc is the grpc.ServiceDesc for Id service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Id_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "id.Id",
	HandlerType: (*IdServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Get",
			Handler:    _Id_Get_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "id.proto",
}
