// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v3.12.4
// source: digimon.proto

package Regionales

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.64.0 or later.
const _ = grpc.SupportPackageIsVersion9

const (
	PrimaryNode_SendStatus_FullMethodName       = "/digimon.PrimaryNode/SendStatus"
	PrimaryNode_FinishRegionales_FullMethodName = "/digimon.PrimaryNode/finishRegionales"
)

// PrimaryNodeClient is the client API for PrimaryNode service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type PrimaryNodeClient interface {
	SendStatus(ctx context.Context, in *DigimonStatus, opts ...grpc.CallOption) (*Response, error)
	FinishRegionales(ctx context.Context, in *FinishRegionalesRequest, opts ...grpc.CallOption) (*FinishRegionalesResponse, error)
}

type primaryNodeClient struct {
	cc grpc.ClientConnInterface
}

func NewPrimaryNodeClient(cc grpc.ClientConnInterface) PrimaryNodeClient {
	return &primaryNodeClient{cc}
}

func (c *primaryNodeClient) SendStatus(ctx context.Context, in *DigimonStatus, opts ...grpc.CallOption) (*Response, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(Response)
	err := c.cc.Invoke(ctx, PrimaryNode_SendStatus_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *primaryNodeClient) FinishRegionales(ctx context.Context, in *FinishRegionalesRequest, opts ...grpc.CallOption) (*FinishRegionalesResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(FinishRegionalesResponse)
	err := c.cc.Invoke(ctx, PrimaryNode_FinishRegionales_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// PrimaryNodeServer is the server API for PrimaryNode service.
// All implementations must embed UnimplementedPrimaryNodeServer
// for forward compatibility.
type PrimaryNodeServer interface {
	SendStatus(context.Context, *DigimonStatus) (*Response, error)
	FinishRegionales(context.Context, *FinishRegionalesRequest) (*FinishRegionalesResponse, error)
	mustEmbedUnimplementedPrimaryNodeServer()
}

// UnimplementedPrimaryNodeServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedPrimaryNodeServer struct{}

func (UnimplementedPrimaryNodeServer) SendStatus(context.Context, *DigimonStatus) (*Response, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SendStatus not implemented")
}
func (UnimplementedPrimaryNodeServer) FinishRegionales(context.Context, *FinishRegionalesRequest) (*FinishRegionalesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method FinishRegionales not implemented")
}
func (UnimplementedPrimaryNodeServer) mustEmbedUnimplementedPrimaryNodeServer() {}
func (UnimplementedPrimaryNodeServer) testEmbeddedByValue()                     {}

// UnsafePrimaryNodeServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to PrimaryNodeServer will
// result in compilation errors.
type UnsafePrimaryNodeServer interface {
	mustEmbedUnimplementedPrimaryNodeServer()
}

func RegisterPrimaryNodeServer(s grpc.ServiceRegistrar, srv PrimaryNodeServer) {
	// If the following call pancis, it indicates UnimplementedPrimaryNodeServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&PrimaryNode_ServiceDesc, srv)
}

func _PrimaryNode_SendStatus_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DigimonStatus)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PrimaryNodeServer).SendStatus(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: PrimaryNode_SendStatus_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PrimaryNodeServer).SendStatus(ctx, req.(*DigimonStatus))
	}
	return interceptor(ctx, in, info, handler)
}

func _PrimaryNode_FinishRegionales_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(FinishRegionalesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PrimaryNodeServer).FinishRegionales(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: PrimaryNode_FinishRegionales_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PrimaryNodeServer).FinishRegionales(ctx, req.(*FinishRegionalesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// PrimaryNode_ServiceDesc is the grpc.ServiceDesc for PrimaryNode service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var PrimaryNode_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "digimon.PrimaryNode",
	HandlerType: (*PrimaryNodeServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "SendStatus",
			Handler:    _PrimaryNode_SendStatus_Handler,
		},
		{
			MethodName: "finishRegionales",
			Handler:    _PrimaryNode_FinishRegionales_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "digimon.proto",
}
