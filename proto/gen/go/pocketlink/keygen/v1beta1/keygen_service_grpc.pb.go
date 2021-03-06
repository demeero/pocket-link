// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package v1beta1

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

// KeygenServiceClient is the client API for KeygenService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type KeygenServiceClient interface {
	GenerateKey(ctx context.Context, in *GenerateKeyRequest, opts ...grpc.CallOption) (*GenerateKeyResponse, error)
}

type keygenServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewKeygenServiceClient(cc grpc.ClientConnInterface) KeygenServiceClient {
	return &keygenServiceClient{cc}
}

func (c *keygenServiceClient) GenerateKey(ctx context.Context, in *GenerateKeyRequest, opts ...grpc.CallOption) (*GenerateKeyResponse, error) {
	out := new(GenerateKeyResponse)
	err := c.cc.Invoke(ctx, "/pocketlink.keygen.v1beta1.KeygenService/GenerateKey", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// KeygenServiceServer is the server API for KeygenService service.
// All implementations must embed UnimplementedKeygenServiceServer
// for forward compatibility
type KeygenServiceServer interface {
	GenerateKey(context.Context, *GenerateKeyRequest) (*GenerateKeyResponse, error)
	mustEmbedUnimplementedKeygenServiceServer()
}

// UnimplementedKeygenServiceServer must be embedded to have forward compatible implementations.
type UnimplementedKeygenServiceServer struct {
}

func (UnimplementedKeygenServiceServer) GenerateKey(context.Context, *GenerateKeyRequest) (*GenerateKeyResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GenerateKey not implemented")
}
func (UnimplementedKeygenServiceServer) mustEmbedUnimplementedKeygenServiceServer() {}

// UnsafeKeygenServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to KeygenServiceServer will
// result in compilation errors.
type UnsafeKeygenServiceServer interface {
	mustEmbedUnimplementedKeygenServiceServer()
}

func RegisterKeygenServiceServer(s grpc.ServiceRegistrar, srv KeygenServiceServer) {
	s.RegisterService(&KeygenService_ServiceDesc, srv)
}

func _KeygenService_GenerateKey_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GenerateKeyRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(KeygenServiceServer).GenerateKey(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/pocketlink.keygen.v1beta1.KeygenService/GenerateKey",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(KeygenServiceServer).GenerateKey(ctx, req.(*GenerateKeyRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// KeygenService_ServiceDesc is the grpc.ServiceDesc for KeygenService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var KeygenService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "pocketlink.keygen.v1beta1.KeygenService",
	HandlerType: (*KeygenServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GenerateKey",
			Handler:    _KeygenService_GenerateKey_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "pocketlink/keygen/v1beta1/keygen_service.proto",
}
