// Code generated by protoc-gen-gogo.
// source: message.proto
// DO NOT EDIT!

/*
	Package blockchain is a generated protocol buffer package.

	It is generated from these files:
		message.proto

	It has these top-level messages:
		BlockRequestMessage
		BlockResponseMessage
		StatusRequestMessage
		StatusResponseMessage
		BlockHeaderRequestMessage
		BlockHeaderResponseMessage
		BlockMessage
*/
package blockchain

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import types "github.com/Baptist-Publication/chorus/angine/protos/types"

import io "io"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type MsgType int32

const (
	MsgType_None      MsgType = 0
	MsgType_BlockReq  MsgType = 1
	MsgType_BlockRsp  MsgType = 2
	MsgType_StatusReq MsgType = 3
	MsgType_StatusRsp MsgType = 4
	MsgType_HeaderReq MsgType = 5
	MsgType_HeaderRsp MsgType = 6
)

var MsgType_name = map[int32]string{
	0: "None",
	1: "BlockReq",
	2: "BlockRsp",
	3: "StatusReq",
	4: "StatusRsp",
	5: "HeaderReq",
	6: "HeaderRsp",
}
var MsgType_value = map[string]int32{
	"None":      0,
	"BlockReq":  1,
	"BlockRsp":  2,
	"StatusReq": 3,
	"StatusRsp": 4,
	"HeaderReq": 5,
	"HeaderRsp": 6,
}

func (x MsgType) String() string {
	return proto.EnumName(MsgType_name, int32(x))
}
func (MsgType) EnumDescriptor() ([]byte, []int) { return fileDescriptorMessage, []int{0} }

type BlockRequestMessage struct {
	Height int64 `protobuf:"varint,1,opt,name=Height,proto3" json:"Height,omitempty"`
}

func (m *BlockRequestMessage) Reset()                    { *m = BlockRequestMessage{} }
func (m *BlockRequestMessage) String() string            { return proto.CompactTextString(m) }
func (*BlockRequestMessage) ProtoMessage()               {}
func (*BlockRequestMessage) Descriptor() ([]byte, []int) { return fileDescriptorMessage, []int{0} }

func (m *BlockRequestMessage) GetHeight() int64 {
	if m != nil {
		return m.Height
	}
	return 0
}

type BlockResponseMessage struct {
	Block *types.Block `protobuf:"bytes,1,opt,name=Block" json:"Block,omitempty"`
}

func (m *BlockResponseMessage) Reset()                    { *m = BlockResponseMessage{} }
func (m *BlockResponseMessage) String() string            { return proto.CompactTextString(m) }
func (*BlockResponseMessage) ProtoMessage()               {}
func (*BlockResponseMessage) Descriptor() ([]byte, []int) { return fileDescriptorMessage, []int{1} }

func (m *BlockResponseMessage) GetBlock() *types.Block {
	if m != nil {
		return m.Block
	}
	return nil
}

type StatusRequestMessage struct {
	Height int64 `protobuf:"varint,1,opt,name=Height,proto3" json:"Height,omitempty"`
}

func (m *StatusRequestMessage) Reset()                    { *m = StatusRequestMessage{} }
func (m *StatusRequestMessage) String() string            { return proto.CompactTextString(m) }
func (*StatusRequestMessage) ProtoMessage()               {}
func (*StatusRequestMessage) Descriptor() ([]byte, []int) { return fileDescriptorMessage, []int{2} }

func (m *StatusRequestMessage) GetHeight() int64 {
	if m != nil {
		return m.Height
	}
	return 0
}

type StatusResponseMessage struct {
	Height int64 `protobuf:"varint,1,opt,name=Height,proto3" json:"Height,omitempty"`
}

func (m *StatusResponseMessage) Reset()                    { *m = StatusResponseMessage{} }
func (m *StatusResponseMessage) String() string            { return proto.CompactTextString(m) }
func (*StatusResponseMessage) ProtoMessage()               {}
func (*StatusResponseMessage) Descriptor() ([]byte, []int) { return fileDescriptorMessage, []int{3} }

func (m *StatusResponseMessage) GetHeight() int64 {
	if m != nil {
		return m.Height
	}
	return 0
}

type BlockHeaderRequestMessage struct {
	Height int64 `protobuf:"varint,1,opt,name=Height,proto3" json:"Height,omitempty"`
}

func (m *BlockHeaderRequestMessage) Reset()                    { *m = BlockHeaderRequestMessage{} }
func (m *BlockHeaderRequestMessage) String() string            { return proto.CompactTextString(m) }
func (*BlockHeaderRequestMessage) ProtoMessage()               {}
func (*BlockHeaderRequestMessage) Descriptor() ([]byte, []int) { return fileDescriptorMessage, []int{4} }

func (m *BlockHeaderRequestMessage) GetHeight() int64 {
	if m != nil {
		return m.Height
	}
	return 0
}

type BlockHeaderResponseMessage struct {
	Header *types.Header `protobuf:"bytes,1,opt,name=Header" json:"Header,omitempty"`
}

func (m *BlockHeaderResponseMessage) Reset()         { *m = BlockHeaderResponseMessage{} }
func (m *BlockHeaderResponseMessage) String() string { return proto.CompactTextString(m) }
func (*BlockHeaderResponseMessage) ProtoMessage()    {}
func (*BlockHeaderResponseMessage) Descriptor() ([]byte, []int) {
	return fileDescriptorMessage, []int{5}
}

func (m *BlockHeaderResponseMessage) GetHeader() *types.Header {
	if m != nil {
		return m.Header
	}
	return nil
}

type BlockMessage struct {
	Type MsgType `protobuf:"varint,1,opt,name=Type,proto3,enum=blockchain.MsgType" json:"Type,omitempty"`
	Data []byte  `protobuf:"bytes,2,opt,name=Data,proto3" json:"Data,omitempty"`
}

func (m *BlockMessage) Reset()                    { *m = BlockMessage{} }
func (m *BlockMessage) String() string            { return proto.CompactTextString(m) }
func (*BlockMessage) ProtoMessage()               {}
func (*BlockMessage) Descriptor() ([]byte, []int) { return fileDescriptorMessage, []int{6} }

func (m *BlockMessage) GetType() MsgType {
	if m != nil {
		return m.Type
	}
	return MsgType_None
}

func (m *BlockMessage) GetData() []byte {
	if m != nil {
		return m.Data
	}
	return nil
}

func init() {
	proto.RegisterType((*BlockRequestMessage)(nil), "blockchain.BlockRequestMessage")
	proto.RegisterType((*BlockResponseMessage)(nil), "blockchain.BlockResponseMessage")
	proto.RegisterType((*StatusRequestMessage)(nil), "blockchain.StatusRequestMessage")
	proto.RegisterType((*StatusResponseMessage)(nil), "blockchain.StatusResponseMessage")
	proto.RegisterType((*BlockHeaderRequestMessage)(nil), "blockchain.BlockHeaderRequestMessage")
	proto.RegisterType((*BlockHeaderResponseMessage)(nil), "blockchain.BlockHeaderResponseMessage")
	proto.RegisterType((*BlockMessage)(nil), "blockchain.BlockMessage")
	proto.RegisterEnum("blockchain.MsgType", MsgType_name, MsgType_value)
}
func (m *BlockRequestMessage) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *BlockRequestMessage) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if m.Height != 0 {
		dAtA[i] = 0x8
		i++
		i = encodeVarintMessage(dAtA, i, uint64(m.Height))
	}
	return i, nil
}

func (m *BlockResponseMessage) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *BlockResponseMessage) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if m.Block != nil {
		dAtA[i] = 0xa
		i++
		i = encodeVarintMessage(dAtA, i, uint64(m.Block.Size()))
		n1, err := m.Block.MarshalTo(dAtA[i:])
		if err != nil {
			return 0, err
		}
		i += n1
	}
	return i, nil
}

func (m *StatusRequestMessage) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *StatusRequestMessage) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if m.Height != 0 {
		dAtA[i] = 0x8
		i++
		i = encodeVarintMessage(dAtA, i, uint64(m.Height))
	}
	return i, nil
}

func (m *StatusResponseMessage) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *StatusResponseMessage) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if m.Height != 0 {
		dAtA[i] = 0x8
		i++
		i = encodeVarintMessage(dAtA, i, uint64(m.Height))
	}
	return i, nil
}

func (m *BlockHeaderRequestMessage) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *BlockHeaderRequestMessage) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if m.Height != 0 {
		dAtA[i] = 0x8
		i++
		i = encodeVarintMessage(dAtA, i, uint64(m.Height))
	}
	return i, nil
}

func (m *BlockHeaderResponseMessage) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *BlockHeaderResponseMessage) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if m.Header != nil {
		dAtA[i] = 0xa
		i++
		i = encodeVarintMessage(dAtA, i, uint64(m.Header.Size()))
		n2, err := m.Header.MarshalTo(dAtA[i:])
		if err != nil {
			return 0, err
		}
		i += n2
	}
	return i, nil
}

func (m *BlockMessage) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *BlockMessage) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if m.Type != 0 {
		dAtA[i] = 0x8
		i++
		i = encodeVarintMessage(dAtA, i, uint64(m.Type))
	}
	if len(m.Data) > 0 {
		dAtA[i] = 0x12
		i++
		i = encodeVarintMessage(dAtA, i, uint64(len(m.Data)))
		i += copy(dAtA[i:], m.Data)
	}
	return i, nil
}

func encodeFixed64Message(dAtA []byte, offset int, v uint64) int {
	dAtA[offset] = uint8(v)
	dAtA[offset+1] = uint8(v >> 8)
	dAtA[offset+2] = uint8(v >> 16)
	dAtA[offset+3] = uint8(v >> 24)
	dAtA[offset+4] = uint8(v >> 32)
	dAtA[offset+5] = uint8(v >> 40)
	dAtA[offset+6] = uint8(v >> 48)
	dAtA[offset+7] = uint8(v >> 56)
	return offset + 8
}
func encodeFixed32Message(dAtA []byte, offset int, v uint32) int {
	dAtA[offset] = uint8(v)
	dAtA[offset+1] = uint8(v >> 8)
	dAtA[offset+2] = uint8(v >> 16)
	dAtA[offset+3] = uint8(v >> 24)
	return offset + 4
}
func encodeVarintMessage(dAtA []byte, offset int, v uint64) int {
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return offset + 1
}
func (m *BlockRequestMessage) Size() (n int) {
	var l int
	_ = l
	if m.Height != 0 {
		n += 1 + sovMessage(uint64(m.Height))
	}
	return n
}

func (m *BlockResponseMessage) Size() (n int) {
	var l int
	_ = l
	if m.Block != nil {
		l = m.Block.Size()
		n += 1 + l + sovMessage(uint64(l))
	}
	return n
}

func (m *StatusRequestMessage) Size() (n int) {
	var l int
	_ = l
	if m.Height != 0 {
		n += 1 + sovMessage(uint64(m.Height))
	}
	return n
}

func (m *StatusResponseMessage) Size() (n int) {
	var l int
	_ = l
	if m.Height != 0 {
		n += 1 + sovMessage(uint64(m.Height))
	}
	return n
}

func (m *BlockHeaderRequestMessage) Size() (n int) {
	var l int
	_ = l
	if m.Height != 0 {
		n += 1 + sovMessage(uint64(m.Height))
	}
	return n
}

func (m *BlockHeaderResponseMessage) Size() (n int) {
	var l int
	_ = l
	if m.Header != nil {
		l = m.Header.Size()
		n += 1 + l + sovMessage(uint64(l))
	}
	return n
}

func (m *BlockMessage) Size() (n int) {
	var l int
	_ = l
	if m.Type != 0 {
		n += 1 + sovMessage(uint64(m.Type))
	}
	l = len(m.Data)
	if l > 0 {
		n += 1 + l + sovMessage(uint64(l))
	}
	return n
}

func sovMessage(x uint64) (n int) {
	for {
		n++
		x >>= 7
		if x == 0 {
			break
		}
	}
	return n
}
func sozMessage(x uint64) (n int) {
	return sovMessage(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *BlockRequestMessage) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowMessage
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: BlockRequestMessage: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: BlockRequestMessage: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Height", wireType)
			}
			m.Height = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMessage
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Height |= (int64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		default:
			iNdEx = preIndex
			skippy, err := skipMessage(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthMessage
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *BlockResponseMessage) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowMessage
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: BlockResponseMessage: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: BlockResponseMessage: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Block", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMessage
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthMessage
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.Block == nil {
				m.Block = &types.Block{}
			}
			if err := m.Block.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipMessage(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthMessage
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *StatusRequestMessage) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowMessage
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: StatusRequestMessage: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: StatusRequestMessage: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Height", wireType)
			}
			m.Height = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMessage
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Height |= (int64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		default:
			iNdEx = preIndex
			skippy, err := skipMessage(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthMessage
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *StatusResponseMessage) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowMessage
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: StatusResponseMessage: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: StatusResponseMessage: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Height", wireType)
			}
			m.Height = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMessage
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Height |= (int64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		default:
			iNdEx = preIndex
			skippy, err := skipMessage(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthMessage
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *BlockHeaderRequestMessage) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowMessage
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: BlockHeaderRequestMessage: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: BlockHeaderRequestMessage: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Height", wireType)
			}
			m.Height = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMessage
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Height |= (int64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		default:
			iNdEx = preIndex
			skippy, err := skipMessage(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthMessage
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *BlockHeaderResponseMessage) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowMessage
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: BlockHeaderResponseMessage: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: BlockHeaderResponseMessage: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Header", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMessage
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthMessage
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.Header == nil {
				m.Header = &types.Header{}
			}
			if err := m.Header.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipMessage(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthMessage
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *BlockMessage) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowMessage
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: BlockMessage: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: BlockMessage: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Type", wireType)
			}
			m.Type = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMessage
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Type |= (MsgType(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Data", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMessage
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return ErrInvalidLengthMessage
			}
			postIndex := iNdEx + byteLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Data = append(m.Data[:0], dAtA[iNdEx:postIndex]...)
			if m.Data == nil {
				m.Data = []byte{}
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipMessage(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthMessage
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func skipMessage(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowMessage
			}
			if iNdEx >= l {
				return 0, io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		wireType := int(wire & 0x7)
		switch wireType {
		case 0:
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowMessage
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
					break
				}
			}
			return iNdEx, nil
		case 1:
			iNdEx += 8
			return iNdEx, nil
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowMessage
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				length |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			iNdEx += length
			if length < 0 {
				return 0, ErrInvalidLengthMessage
			}
			return iNdEx, nil
		case 3:
			for {
				var innerWire uint64
				var start int = iNdEx
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return 0, ErrIntOverflowMessage
					}
					if iNdEx >= l {
						return 0, io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					innerWire |= (uint64(b) & 0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				innerWireType := int(innerWire & 0x7)
				if innerWireType == 4 {
					break
				}
				next, err := skipMessage(dAtA[start:])
				if err != nil {
					return 0, err
				}
				iNdEx = start + next
			}
			return iNdEx, nil
		case 4:
			return iNdEx, nil
		case 5:
			iNdEx += 4
			return iNdEx, nil
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
	}
	panic("unreachable")
}

var (
	ErrInvalidLengthMessage = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowMessage   = fmt.Errorf("proto: integer overflow")
)

func init() { proto.RegisterFile("message.proto", fileDescriptorMessage) }

var fileDescriptorMessage = []byte{
	// 347 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x8c, 0x92, 0x4f, 0x6e, 0xe2, 0x30,
	0x18, 0xc5, 0x31, 0x04, 0x86, 0xf9, 0x26, 0x8c, 0x2c, 0xc3, 0x8c, 0x28, 0x8b, 0x08, 0x45, 0xaa,
	0x8a, 0x2a, 0xd5, 0x91, 0x60, 0x57, 0x75, 0x45, 0xbb, 0x40, 0xaa, 0xe8, 0x22, 0xed, 0x05, 0x0c,
	0xb5, 0x42, 0x04, 0xd8, 0x2e, 0x36, 0x0b, 0x7a, 0x92, 0x1e, 0xa9, 0xcb, 0x1e, 0xa1, 0xa2, 0x17,
	0xa9, 0xe2, 0x38, 0xfc, 0x59, 0x54, 0x62, 0x13, 0xe5, 0xe5, 0xfd, 0xde, 0xf7, 0x64, 0x7f, 0x81,
	0xc6, 0x92, 0x6b, 0xcd, 0x12, 0x4e, 0xd5, 0x4a, 0x1a, 0x49, 0x60, 0xb2, 0x90, 0xd3, 0xf9, 0x74,
	0xc6, 0x52, 0xd1, 0xb9, 0x49, 0x52, 0xb3, 0x60, 0x13, 0xfa, 0x3a, 0x93, 0x22, 0x61, 0x42, 0x8a,
	0x45, 0x2a, 0x38, 0x9d, 0xca, 0x65, 0xc4, 0x84, 0x88, 0x98, 0x48, 0x52, 0xc1, 0x23, 0x1b, 0xd3,
	0x91, 0xd9, 0x28, 0xee, 0x9e, 0xf9, 0xa4, 0xf0, 0x0a, 0x9a, 0xc3, 0x6c, 0x56, 0xcc, 0x5f, 0xd6,
	0x5c, 0x9b, 0x71, 0x5e, 0x43, 0xfe, 0x43, 0x6d, 0xc4, 0xd3, 0x64, 0x66, 0xda, 0xa8, 0x8b, 0x7a,
	0x95, 0xd8, 0xa9, 0xf0, 0x1a, 0x5a, 0x0e, 0xd7, 0x4a, 0x0a, 0xcd, 0x0b, 0x3e, 0x84, 0xaa, 0xfd,
	0x6e, 0xf1, 0x3f, 0x7d, 0x9f, 0xe6, 0x1d, 0x39, 0x9b, 0x5b, 0x21, 0x85, 0xd6, 0xa3, 0x61, 0x66,
	0xad, 0x4f, 0xec, 0x8a, 0xe0, 0x5f, 0xc1, 0x1f, 0x97, 0xfd, 0x14, 0x18, 0xc0, 0x99, 0x6d, 0x1a,
	0x71, 0xf6, 0xcc, 0x57, 0x27, 0xb6, 0xdc, 0x42, 0xe7, 0x28, 0x74, 0x5c, 0x75, 0x9e, 0xa5, 0x32,
	0xc3, 0x1d, 0xac, 0xe1, 0x0e, 0xe6, 0x68, 0x67, 0x86, 0xf7, 0xe0, 0xdb, 0x21, 0x45, 0xec, 0x02,
	0xbc, 0xa7, 0x8d, 0xe2, 0x36, 0xf4, 0xb7, 0xdf, 0xa4, 0xfb, 0x75, 0xd1, 0xb1, 0x4e, 0x32, 0x2b,
	0xb6, 0x00, 0x21, 0xe0, 0xdd, 0x31, 0xc3, 0xda, 0xe5, 0x2e, 0xea, 0xf9, 0xb1, 0x7d, 0xbf, 0x9c,
	0xc3, 0x2f, 0x07, 0x91, 0x3a, 0x78, 0x0f, 0x52, 0x70, 0x5c, 0x22, 0x3e, 0xd4, 0x8b, 0x3d, 0x61,
	0xb4, 0x57, 0x5a, 0xe1, 0x32, 0x69, 0xc0, 0xef, 0xdd, 0xc5, 0xe2, 0xca, 0x81, 0xd4, 0x0a, 0x7b,
	0x99, 0xdc, 0x5d, 0x08, 0xae, 0x1e, 0x48, 0xad, 0x70, 0x6d, 0x88, 0xdf, 0xb7, 0x01, 0xfa, 0xd8,
	0x06, 0xe8, 0x73, 0x1b, 0xa0, 0xb7, 0xaf, 0xa0, 0x34, 0xa9, 0xd9, 0x1f, 0x63, 0xf0, 0x1d, 0x00,
	0x00, 0xff, 0xff, 0x05, 0xc3, 0x6d, 0xff, 0x73, 0x02, 0x00, 0x00,
}
