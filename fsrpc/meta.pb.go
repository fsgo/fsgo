// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.29.0-devel
// 	protoc        v3.21.12
// source: meta.proto

package fsrpc

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	anypb "google.golang.org/protobuf/types/known/anypb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// CompressType 数据压缩类型
type CompressType int32

const (
	CompressType_No   CompressType = 0
	CompressType_GZIP CompressType = 1
)

// Enum value maps for CompressType.
var (
	CompressType_name = map[int32]string{
		0: "No",
		1: "GZIP",
	}
	CompressType_value = map[string]int32{
		"No":   0,
		"GZIP": 1,
	}
)

func (x CompressType) Enum() *CompressType {
	p := new(CompressType)
	*p = x
	return p
}

func (x CompressType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (CompressType) Descriptor() protoreflect.EnumDescriptor {
	return file_meta_proto_enumTypes[0].Descriptor()
}

func (CompressType) Type() protoreflect.EnumType {
	return &file_meta_proto_enumTypes[0]
}

func (x CompressType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use CompressType.Descriptor instead.
func (CompressType) EnumDescriptor() ([]byte, []int) {
	return file_meta_proto_rawDescGZIP(), []int{0}
}

// 错误编码
type ErrCode int32

const (
	ErrCode_Success                ErrCode = 0    // 成功
	ErrCode_ReqNoAuth              ErrCode = 1001 // 没有校验权限
	ErrCode_ReqNoService           ErrCode = 1002 // service 不存在
	ErrCode_ReqNoMethod            ErrCode = 1003 // method 不存在
	ErrCode_ReqUnknownCompressType ErrCode = 1004 // 不支持的压缩类型
	ErrCode_ReqBadParams           ErrCode = 1005 // 参数异常
	ErrCode_SerInternal            ErrCode = 2001 // 内部错误
	ErrCode_SerShutdown            ErrCode = 2002 // 服务正在关闭
	ErrCode_SerLimit               ErrCode = 2003 // 服务达到处理上线，当前请求被限流
	ErrCode_SerBadConn             ErrCode = 2004 // 异常的网络连接
)

// Enum value maps for ErrCode.
var (
	ErrCode_name = map[int32]string{
		0:    "Success",
		1001: "ReqNoAuth",
		1002: "ReqNoService",
		1003: "ReqNoMethod",
		1004: "ReqUnknownCompressType",
		1005: "ReqBadParams",
		2001: "SerInternal",
		2002: "SerShutdown",
		2003: "SerLimit",
		2004: "SerBadConn",
	}
	ErrCode_value = map[string]int32{
		"Success":                0,
		"ReqNoAuth":              1001,
		"ReqNoService":           1002,
		"ReqNoMethod":            1003,
		"ReqUnknownCompressType": 1004,
		"ReqBadParams":           1005,
		"SerInternal":            2001,
		"SerShutdown":            2002,
		"SerLimit":               2003,
		"SerBadConn":             2004,
	}
)

func (x ErrCode) Enum() *ErrCode {
	p := new(ErrCode)
	*p = x
	return p
}

func (x ErrCode) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (ErrCode) Descriptor() protoreflect.EnumDescriptor {
	return file_meta_proto_enumTypes[1].Descriptor()
}

func (ErrCode) Type() protoreflect.EnumType {
	return &file_meta_proto_enumTypes[1]
}

func (x ErrCode) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use ErrCode.Descriptor instead.
func (ErrCode) EnumDescriptor() ([]byte, []int) {
	return file_meta_proto_rawDescGZIP(), []int{1}
}

// Request 一条请求信息
type Request struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Method       string                `protobuf:"bytes,1,opt,name=Method,proto3" json:"Method,omitempty"`                                      // 请求的方法
	ID           uint64                `protobuf:"varint,2,opt,name=ID,proto3" json:"ID,omitempty"`                                             // 请求的唯一 ID
	HasPayload   bool                  `protobuf:"varint,3,opt,name=HasPayload,proto3" json:"HasPayload,omitempty"`                             // 是否有消息体
	CompressType CompressType          `protobuf:"varint,4,opt,name=CompressType,proto3,enum=fsrpc.CompressType" json:"CompressType,omitempty"` // 数据压缩类型
	LogID        string                `protobuf:"bytes,5,opt,name=LogID,proto3" json:"LogID,omitempty"`                                        // 日志 ID
	TraceID      string                `protobuf:"bytes,6,opt,name=TraceID,proto3" json:"TraceID,omitempty"`
	SpanID       string                `protobuf:"bytes,7,opt,name=SpanID,proto3" json:"SpanID,omitempty"`
	ParentSpanID string                `protobuf:"bytes,8,opt,name=ParentSpanID,proto3" json:"ParentSpanID,omitempty"`
	ExtKV        map[string]*anypb.Any `protobuf:"bytes,9,rep,name=ExtKV,proto3" json:"ExtKV,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"` // 其他 kv 数据
}

func (x *Request) Reset() {
	*x = Request{}
	if protoimpl.UnsafeEnabled {
		mi := &file_meta_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Request) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Request) ProtoMessage() {}

func (x *Request) ProtoReflect() protoreflect.Message {
	mi := &file_meta_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Request.ProtoReflect.Descriptor instead.
func (*Request) Descriptor() ([]byte, []int) {
	return file_meta_proto_rawDescGZIP(), []int{0}
}

func (x *Request) GetMethod() string {
	if x != nil {
		return x.Method
	}
	return ""
}

func (x *Request) GetID() uint64 {
	if x != nil {
		return x.ID
	}
	return 0
}

func (x *Request) GetHasPayload() bool {
	if x != nil {
		return x.HasPayload
	}
	return false
}

func (x *Request) GetCompressType() CompressType {
	if x != nil {
		return x.CompressType
	}
	return CompressType_No
}

func (x *Request) GetLogID() string {
	if x != nil {
		return x.LogID
	}
	return ""
}

func (x *Request) GetTraceID() string {
	if x != nil {
		return x.TraceID
	}
	return ""
}

func (x *Request) GetSpanID() string {
	if x != nil {
		return x.SpanID
	}
	return ""
}

func (x *Request) GetParentSpanID() string {
	if x != nil {
		return x.ParentSpanID
	}
	return ""
}

func (x *Request) GetExtKV() map[string]*anypb.Any {
	if x != nil {
		return x.ExtKV
	}
	return nil
}

type Payload struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Index  uint32 `protobuf:"varint,1,opt,name=Index,proto3" json:"Index,omitempty"`   // 数据的编号，从 0 依次递增
	RID    uint64 `protobuf:"varint,2,opt,name=RID,proto3" json:"RID,omitempty"`       // Request 或者  Response 的 ID
	More   bool   `protobuf:"varint,3,opt,name=More,proto3" json:"More,omitempty"`     // 是否有更多消息
	Length uint32 `protobuf:"varint,4,opt,name=Length,proto3" json:"Length,omitempty"` // 数据内容长度
}

func (x *Payload) Reset() {
	*x = Payload{}
	if protoimpl.UnsafeEnabled {
		mi := &file_meta_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Payload) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Payload) ProtoMessage() {}

func (x *Payload) ProtoReflect() protoreflect.Message {
	mi := &file_meta_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Payload.ProtoReflect.Descriptor instead.
func (*Payload) Descriptor() ([]byte, []int) {
	return file_meta_proto_rawDescGZIP(), []int{1}
}

func (x *Payload) GetIndex() uint32 {
	if x != nil {
		return x.Index
	}
	return 0
}

func (x *Payload) GetRID() uint64 {
	if x != nil {
		return x.RID
	}
	return 0
}

func (x *Payload) GetMore() bool {
	if x != nil {
		return x.More
	}
	return false
}

func (x *Payload) GetLength() uint32 {
	if x != nil {
		return x.Length
	}
	return 0
}

type Response struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Code         ErrCode               `protobuf:"varint,1,opt,name=Code,proto3,enum=fsrpc.ErrCode" json:"Code,omitempty"`                                                                       // 错误码
	Message      string                `protobuf:"bytes,2,opt,name=Message,proto3" json:"Message,omitempty"`                                                                                     // 错误详情
	RequestID    uint64                `protobuf:"varint,3,opt,name=RequestID,proto3" json:"RequestID,omitempty"`                                                                                // 请求的唯一 ID
	HasPayload   bool                  `protobuf:"varint,4,opt,name=HasPayload,proto3" json:"HasPayload,omitempty"`                                                                              // 是否有消息体
	CompressType CompressType          `protobuf:"varint,5,opt,name=CompressType,proto3,enum=fsrpc.CompressType" json:"CompressType,omitempty"`                                                  // 消息体压缩类型
	ExtKV        map[string]*anypb.Any `protobuf:"bytes,6,rep,name=ExtKV,proto3" json:"ExtKV,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"` // 其他 kv 数据
}

func (x *Response) Reset() {
	*x = Response{}
	if protoimpl.UnsafeEnabled {
		mi := &file_meta_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Response) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Response) ProtoMessage() {}

func (x *Response) ProtoReflect() protoreflect.Message {
	mi := &file_meta_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Response.ProtoReflect.Descriptor instead.
func (*Response) Descriptor() ([]byte, []int) {
	return file_meta_proto_rawDescGZIP(), []int{2}
}

func (x *Response) GetCode() ErrCode {
	if x != nil {
		return x.Code
	}
	return ErrCode_Success
}

func (x *Response) GetMessage() string {
	if x != nil {
		return x.Message
	}
	return ""
}

func (x *Response) GetRequestID() uint64 {
	if x != nil {
		return x.RequestID
	}
	return 0
}

func (x *Response) GetHasPayload() bool {
	if x != nil {
		return x.HasPayload
	}
	return false
}

func (x *Response) GetCompressType() CompressType {
	if x != nil {
		return x.CompressType
	}
	return CompressType_No
}

func (x *Response) GetExtKV() map[string]*anypb.Any {
	if x != nil {
		return x.ExtKV
	}
	return nil
}

// AuthRequest 鉴权请求
type AuthRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	UserName string `protobuf:"bytes,1,opt,name=UserName,proto3" json:"UserName,omitempty"`  // 用户名
	Token    string `protobuf:"bytes,2,opt,name=Token,proto3" json:"Token,omitempty"`        // 鉴权的密码
	Timespan int64  `protobuf:"varint,3,opt,name=Timespan,proto3" json:"Timespan,omitempty"` // 当前时间错，单位-秒
	Type     string `protobuf:"bytes,4,opt,name=Type,proto3" json:"Type,omitempty"`          // 鉴权算法类型，可选
}

func (x *AuthRequest) Reset() {
	*x = AuthRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_meta_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AuthRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AuthRequest) ProtoMessage() {}

func (x *AuthRequest) ProtoReflect() protoreflect.Message {
	mi := &file_meta_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AuthRequest.ProtoReflect.Descriptor instead.
func (*AuthRequest) Descriptor() ([]byte, []int) {
	return file_meta_proto_rawDescGZIP(), []int{3}
}

func (x *AuthRequest) GetUserName() string {
	if x != nil {
		return x.UserName
	}
	return ""
}

func (x *AuthRequest) GetToken() string {
	if x != nil {
		return x.Token
	}
	return ""
}

func (x *AuthRequest) GetTimespan() int64 {
	if x != nil {
		return x.Timespan
	}
	return 0
}

func (x *AuthRequest) GetType() string {
	if x != nil {
		return x.Type
	}
	return ""
}

type PingPong struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ID      uint64 `protobuf:"varint,1,opt,name=ID,proto3" json:"ID,omitempty"`          // 编号，发送 Ping 消息 ID=1，则回复时 ID=2
	Message string `protobuf:"bytes,2,opt,name=Message,proto3" json:"Message,omitempty"` // 消息内容
}

func (x *PingPong) Reset() {
	*x = PingPong{}
	if protoimpl.UnsafeEnabled {
		mi := &file_meta_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PingPong) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PingPong) ProtoMessage() {}

func (x *PingPong) ProtoReflect() protoreflect.Message {
	mi := &file_meta_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PingPong.ProtoReflect.Descriptor instead.
func (*PingPong) Descriptor() ([]byte, []int) {
	return file_meta_proto_rawDescGZIP(), []int{4}
}

func (x *PingPong) GetID() uint64 {
	if x != nil {
		return x.ID
	}
	return 0
}

func (x *PingPong) GetMessage() string {
	if x != nil {
		return x.Message
	}
	return ""
}

var File_meta_proto protoreflect.FileDescriptor

var file_meta_proto_rawDesc = []byte{
	0x0a, 0x0a, 0x6d, 0x65, 0x74, 0x61, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x05, 0x66, 0x73,
	0x72, 0x70, 0x63, 0x1a, 0x19, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x62, 0x75, 0x66, 0x2f, 0x61, 0x6e, 0x79, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xf7,
	0x02, 0x0a, 0x07, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x16, 0x0a, 0x06, 0x4d, 0x65,
	0x74, 0x68, 0x6f, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x4d, 0x65, 0x74, 0x68,
	0x6f, 0x64, 0x12, 0x0e, 0x0a, 0x02, 0x49, 0x44, 0x18, 0x02, 0x20, 0x01, 0x28, 0x04, 0x52, 0x02,
	0x49, 0x44, 0x12, 0x1e, 0x0a, 0x0a, 0x48, 0x61, 0x73, 0x50, 0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64,
	0x18, 0x03, 0x20, 0x01, 0x28, 0x08, 0x52, 0x0a, 0x48, 0x61, 0x73, 0x50, 0x61, 0x79, 0x6c, 0x6f,
	0x61, 0x64, 0x12, 0x37, 0x0a, 0x0c, 0x43, 0x6f, 0x6d, 0x70, 0x72, 0x65, 0x73, 0x73, 0x54, 0x79,
	0x70, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x13, 0x2e, 0x66, 0x73, 0x72, 0x70, 0x63,
	0x2e, 0x43, 0x6f, 0x6d, 0x70, 0x72, 0x65, 0x73, 0x73, 0x54, 0x79, 0x70, 0x65, 0x52, 0x0c, 0x43,
	0x6f, 0x6d, 0x70, 0x72, 0x65, 0x73, 0x73, 0x54, 0x79, 0x70, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x4c,
	0x6f, 0x67, 0x49, 0x44, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x4c, 0x6f, 0x67, 0x49,
	0x44, 0x12, 0x18, 0x0a, 0x07, 0x54, 0x72, 0x61, 0x63, 0x65, 0x49, 0x44, 0x18, 0x06, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x07, 0x54, 0x72, 0x61, 0x63, 0x65, 0x49, 0x44, 0x12, 0x16, 0x0a, 0x06, 0x53,
	0x70, 0x61, 0x6e, 0x49, 0x44, 0x18, 0x07, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x53, 0x70, 0x61,
	0x6e, 0x49, 0x44, 0x12, 0x22, 0x0a, 0x0c, 0x50, 0x61, 0x72, 0x65, 0x6e, 0x74, 0x53, 0x70, 0x61,
	0x6e, 0x49, 0x44, 0x18, 0x08, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0c, 0x50, 0x61, 0x72, 0x65, 0x6e,
	0x74, 0x53, 0x70, 0x61, 0x6e, 0x49, 0x44, 0x12, 0x2f, 0x0a, 0x05, 0x45, 0x78, 0x74, 0x4b, 0x56,
	0x18, 0x09, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x19, 0x2e, 0x66, 0x73, 0x72, 0x70, 0x63, 0x2e, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x2e, 0x45, 0x78, 0x74, 0x4b, 0x56, 0x45, 0x6e, 0x74, 0x72,
	0x79, 0x52, 0x05, 0x45, 0x78, 0x74, 0x4b, 0x56, 0x1a, 0x4e, 0x0a, 0x0a, 0x45, 0x78, 0x74, 0x4b,
	0x56, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x2a, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75,
	0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x14, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x41, 0x6e, 0x79, 0x52, 0x05, 0x76,
	0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x22, 0x5d, 0x0a, 0x07, 0x50, 0x61, 0x79, 0x6c,
	0x6f, 0x61, 0x64, 0x12, 0x14, 0x0a, 0x05, 0x49, 0x6e, 0x64, 0x65, 0x78, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x0d, 0x52, 0x05, 0x49, 0x6e, 0x64, 0x65, 0x78, 0x12, 0x10, 0x0a, 0x03, 0x52, 0x49, 0x44,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x04, 0x52, 0x03, 0x52, 0x49, 0x44, 0x12, 0x12, 0x0a, 0x04, 0x4d,
	0x6f, 0x72, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x08, 0x52, 0x04, 0x4d, 0x6f, 0x72, 0x65, 0x12,
	0x16, 0x0a, 0x06, 0x4c, 0x65, 0x6e, 0x67, 0x74, 0x68, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0d, 0x52,
	0x06, 0x4c, 0x65, 0x6e, 0x67, 0x74, 0x68, 0x22, 0xc1, 0x02, 0x0a, 0x08, 0x52, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x12, 0x22, 0x0a, 0x04, 0x43, 0x6f, 0x64, 0x65, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x0e, 0x32, 0x0e, 0x2e, 0x66, 0x73, 0x72, 0x70, 0x63, 0x2e, 0x45, 0x72, 0x72, 0x43, 0x6f,
	0x64, 0x65, 0x52, 0x04, 0x43, 0x6f, 0x64, 0x65, 0x12, 0x18, 0x0a, 0x07, 0x4d, 0x65, 0x73, 0x73,
	0x61, 0x67, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x4d, 0x65, 0x73, 0x73, 0x61,
	0x67, 0x65, 0x12, 0x1c, 0x0a, 0x09, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x49, 0x44, 0x18,
	0x03, 0x20, 0x01, 0x28, 0x04, 0x52, 0x09, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x49, 0x44,
	0x12, 0x1e, 0x0a, 0x0a, 0x48, 0x61, 0x73, 0x50, 0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64, 0x18, 0x04,
	0x20, 0x01, 0x28, 0x08, 0x52, 0x0a, 0x48, 0x61, 0x73, 0x50, 0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64,
	0x12, 0x37, 0x0a, 0x0c, 0x43, 0x6f, 0x6d, 0x70, 0x72, 0x65, 0x73, 0x73, 0x54, 0x79, 0x70, 0x65,
	0x18, 0x05, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x13, 0x2e, 0x66, 0x73, 0x72, 0x70, 0x63, 0x2e, 0x43,
	0x6f, 0x6d, 0x70, 0x72, 0x65, 0x73, 0x73, 0x54, 0x79, 0x70, 0x65, 0x52, 0x0c, 0x43, 0x6f, 0x6d,
	0x70, 0x72, 0x65, 0x73, 0x73, 0x54, 0x79, 0x70, 0x65, 0x12, 0x30, 0x0a, 0x05, 0x45, 0x78, 0x74,
	0x4b, 0x56, 0x18, 0x06, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x66, 0x73, 0x72, 0x70, 0x63,
	0x2e, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x2e, 0x45, 0x78, 0x74, 0x4b, 0x56, 0x45,
	0x6e, 0x74, 0x72, 0x79, 0x52, 0x05, 0x45, 0x78, 0x74, 0x4b, 0x56, 0x1a, 0x4e, 0x0a, 0x0a, 0x45,
	0x78, 0x74, 0x4b, 0x56, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x2a, 0x0a, 0x05, 0x76,
	0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x14, 0x2e, 0x67, 0x6f, 0x6f,
	0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x41, 0x6e, 0x79,
	0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x22, 0x6f, 0x0a, 0x0b, 0x41,
	0x75, 0x74, 0x68, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x1a, 0x0a, 0x08, 0x55, 0x73,
	0x65, 0x72, 0x4e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x55, 0x73,
	0x65, 0x72, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x54, 0x6f, 0x6b, 0x65, 0x6e, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x54, 0x6f, 0x6b, 0x65, 0x6e, 0x12, 0x1a, 0x0a, 0x08,
	0x54, 0x69, 0x6d, 0x65, 0x73, 0x70, 0x61, 0x6e, 0x18, 0x03, 0x20, 0x01, 0x28, 0x03, 0x52, 0x08,
	0x54, 0x69, 0x6d, 0x65, 0x73, 0x70, 0x61, 0x6e, 0x12, 0x12, 0x0a, 0x04, 0x54, 0x79, 0x70, 0x65,
	0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x54, 0x79, 0x70, 0x65, 0x22, 0x34, 0x0a, 0x08,
	0x50, 0x69, 0x6e, 0x67, 0x50, 0x6f, 0x6e, 0x67, 0x12, 0x0e, 0x0a, 0x02, 0x49, 0x44, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x04, 0x52, 0x02, 0x49, 0x44, 0x12, 0x18, 0x0a, 0x07, 0x4d, 0x65, 0x73, 0x73,
	0x61, 0x67, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x4d, 0x65, 0x73, 0x73, 0x61,
	0x67, 0x65, 0x2a, 0x20, 0x0a, 0x0c, 0x43, 0x6f, 0x6d, 0x70, 0x72, 0x65, 0x73, 0x73, 0x54, 0x79,
	0x70, 0x65, 0x12, 0x06, 0x0a, 0x02, 0x4e, 0x6f, 0x10, 0x00, 0x12, 0x08, 0x0a, 0x04, 0x47, 0x5a,
	0x49, 0x50, 0x10, 0x01, 0x2a, 0xbf, 0x01, 0x0a, 0x07, 0x45, 0x72, 0x72, 0x43, 0x6f, 0x64, 0x65,
	0x12, 0x0b, 0x0a, 0x07, 0x53, 0x75, 0x63, 0x63, 0x65, 0x73, 0x73, 0x10, 0x00, 0x12, 0x0e, 0x0a,
	0x09, 0x52, 0x65, 0x71, 0x4e, 0x6f, 0x41, 0x75, 0x74, 0x68, 0x10, 0xe9, 0x07, 0x12, 0x11, 0x0a,
	0x0c, 0x52, 0x65, 0x71, 0x4e, 0x6f, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x10, 0xea, 0x07,
	0x12, 0x10, 0x0a, 0x0b, 0x52, 0x65, 0x71, 0x4e, 0x6f, 0x4d, 0x65, 0x74, 0x68, 0x6f, 0x64, 0x10,
	0xeb, 0x07, 0x12, 0x1b, 0x0a, 0x16, 0x52, 0x65, 0x71, 0x55, 0x6e, 0x6b, 0x6e, 0x6f, 0x77, 0x6e,
	0x43, 0x6f, 0x6d, 0x70, 0x72, 0x65, 0x73, 0x73, 0x54, 0x79, 0x70, 0x65, 0x10, 0xec, 0x07, 0x12,
	0x11, 0x0a, 0x0c, 0x52, 0x65, 0x71, 0x42, 0x61, 0x64, 0x50, 0x61, 0x72, 0x61, 0x6d, 0x73, 0x10,
	0xed, 0x07, 0x12, 0x10, 0x0a, 0x0b, 0x53, 0x65, 0x72, 0x49, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61,
	0x6c, 0x10, 0xd1, 0x0f, 0x12, 0x10, 0x0a, 0x0b, 0x53, 0x65, 0x72, 0x53, 0x68, 0x75, 0x74, 0x64,
	0x6f, 0x77, 0x6e, 0x10, 0xd2, 0x0f, 0x12, 0x0d, 0x0a, 0x08, 0x53, 0x65, 0x72, 0x4c, 0x69, 0x6d,
	0x69, 0x74, 0x10, 0xd3, 0x0f, 0x12, 0x0f, 0x0a, 0x0a, 0x53, 0x65, 0x72, 0x42, 0x61, 0x64, 0x43,
	0x6f, 0x6e, 0x6e, 0x10, 0xd4, 0x0f, 0x42, 0x0a, 0x5a, 0x08, 0x2e, 0x2e, 0x2f, 0x66, 0x73, 0x72,
	0x70, 0x63, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_meta_proto_rawDescOnce sync.Once
	file_meta_proto_rawDescData = file_meta_proto_rawDesc
)

func file_meta_proto_rawDescGZIP() []byte {
	file_meta_proto_rawDescOnce.Do(func() {
		file_meta_proto_rawDescData = protoimpl.X.CompressGZIP(file_meta_proto_rawDescData)
	})
	return file_meta_proto_rawDescData
}

var file_meta_proto_enumTypes = make([]protoimpl.EnumInfo, 2)
var file_meta_proto_msgTypes = make([]protoimpl.MessageInfo, 7)
var file_meta_proto_goTypes = []interface{}{
	(CompressType)(0),   // 0: fsrpc.CompressType
	(ErrCode)(0),        // 1: fsrpc.ErrCode
	(*Request)(nil),     // 2: fsrpc.Request
	(*Payload)(nil),     // 3: fsrpc.Payload
	(*Response)(nil),    // 4: fsrpc.Response
	(*AuthRequest)(nil), // 5: fsrpc.AuthRequest
	(*PingPong)(nil),    // 6: fsrpc.PingPong
	nil,                 // 7: fsrpc.Request.ExtKVEntry
	nil,                 // 8: fsrpc.Response.ExtKVEntry
	(*anypb.Any)(nil),   // 9: google.protobuf.Any
}
var file_meta_proto_depIdxs = []int32{
	0, // 0: fsrpc.Request.CompressType:type_name -> fsrpc.CompressType
	7, // 1: fsrpc.Request.ExtKV:type_name -> fsrpc.Request.ExtKVEntry
	1, // 2: fsrpc.Response.Code:type_name -> fsrpc.ErrCode
	0, // 3: fsrpc.Response.CompressType:type_name -> fsrpc.CompressType
	8, // 4: fsrpc.Response.ExtKV:type_name -> fsrpc.Response.ExtKVEntry
	9, // 5: fsrpc.Request.ExtKVEntry.value:type_name -> google.protobuf.Any
	9, // 6: fsrpc.Response.ExtKVEntry.value:type_name -> google.protobuf.Any
	7, // [7:7] is the sub-list for method output_type
	7, // [7:7] is the sub-list for method input_type
	7, // [7:7] is the sub-list for extension type_name
	7, // [7:7] is the sub-list for extension extendee
	0, // [0:7] is the sub-list for field type_name
}

func init() { file_meta_proto_init() }
func file_meta_proto_init() {
	if File_meta_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_meta_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Request); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_meta_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Payload); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_meta_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Response); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_meta_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*AuthRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_meta_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PingPong); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_meta_proto_rawDesc,
			NumEnums:      2,
			NumMessages:   7,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_meta_proto_goTypes,
		DependencyIndexes: file_meta_proto_depIdxs,
		EnumInfos:         file_meta_proto_enumTypes,
		MessageInfos:      file_meta_proto_msgTypes,
	}.Build()
	File_meta_proto = out.File
	file_meta_proto_rawDesc = nil
	file_meta_proto_goTypes = nil
	file_meta_proto_depIdxs = nil
}
