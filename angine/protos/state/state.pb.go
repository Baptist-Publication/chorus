// Code generated by protoc-gen-gogo.
// source: state.proto
// DO NOT EDIT!

/*
	Package state is a generated protocol buffer package.

	It is generated from these files:
		state.proto

	It has these top-level messages:
		GenesisDoc
		ValidatorSet
		SpecialOp
		SuspectPlugin
		QueryCachePlugin
		Plugin
		State
*/
package state

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

type Type int32

const (
	Type_PluginNone       Type = 0
	Type_PluginSpecialOp  Type = 1
	Type_PluginSuspect    Type = 2
	Type_PluginQueryCache Type = 3
)

var Type_name = map[int32]string{
	0: "PluginNone",
	1: "PluginSpecialOp",
	2: "PluginSuspect",
	3: "PluginQueryCache",
}
var Type_value = map[string]int32{
	"PluginNone":       0,
	"PluginSpecialOp":  1,
	"PluginSuspect":    2,
	"PluginQueryCache": 3,
}

func (x Type) String() string {
	return proto.EnumName(Type_name, int32(x))
}
func (Type) EnumDescriptor() ([]byte, []int) { return fileDescriptorState, []int{0} }

type GenesisDoc struct {
	JSONData []byte `protobuf:"bytes,1,opt,name=JSONData,proto3" json:"JSONData,omitempty"`
}

func (m *GenesisDoc) Reset()                    { *m = GenesisDoc{} }
func (m *GenesisDoc) String() string            { return proto.CompactTextString(m) }
func (*GenesisDoc) ProtoMessage()               {}
func (*GenesisDoc) Descriptor() ([]byte, []int) { return fileDescriptorState, []int{0} }

func (m *GenesisDoc) GetJSONData() []byte {
	if m != nil {
		return m.JSONData
	}
	return nil
}

type ValidatorSet struct {
	JSONData []byte `protobuf:"bytes,2,opt,name=JSONData,proto3" json:"JSONData,omitempty"`
}

func (m *ValidatorSet) Reset()                    { *m = ValidatorSet{} }
func (m *ValidatorSet) String() string            { return proto.CompactTextString(m) }
func (*ValidatorSet) ProtoMessage()               {}
func (*ValidatorSet) Descriptor() ([]byte, []int) { return fileDescriptorState, []int{1} }

func (m *ValidatorSet) GetJSONData() []byte {
	if m != nil {
		return m.JSONData
	}
	return nil
}

type SpecialOp struct {
	JSONData []byte `protobuf:"bytes,1,opt,name=JSONData,proto3" json:"JSONData,omitempty"`
}

func (m *SpecialOp) Reset()                    { *m = SpecialOp{} }
func (m *SpecialOp) String() string            { return proto.CompactTextString(m) }
func (*SpecialOp) ProtoMessage()               {}
func (*SpecialOp) Descriptor() ([]byte, []int) { return fileDescriptorState, []int{2} }

func (m *SpecialOp) GetJSONData() []byte {
	if m != nil {
		return m.JSONData
	}
	return nil
}

type SuspectPlugin struct {
	JSONData []byte `protobuf:"bytes,1,opt,name=JSONData,proto3" json:"JSONData,omitempty"`
}

func (m *SuspectPlugin) Reset()                    { *m = SuspectPlugin{} }
func (m *SuspectPlugin) String() string            { return proto.CompactTextString(m) }
func (*SuspectPlugin) ProtoMessage()               {}
func (*SuspectPlugin) Descriptor() ([]byte, []int) { return fileDescriptorState, []int{3} }

func (m *SuspectPlugin) GetJSONData() []byte {
	if m != nil {
		return m.JSONData
	}
	return nil
}

type QueryCachePlugin struct {
	JSONData []byte `protobuf:"bytes,1,opt,name=JSONData,proto3" json:"JSONData,omitempty"`
}

func (m *QueryCachePlugin) Reset()                    { *m = QueryCachePlugin{} }
func (m *QueryCachePlugin) String() string            { return proto.CompactTextString(m) }
func (*QueryCachePlugin) ProtoMessage()               {}
func (*QueryCachePlugin) Descriptor() ([]byte, []int) { return fileDescriptorState, []int{4} }

func (m *QueryCachePlugin) GetJSONData() []byte {
	if m != nil {
		return m.JSONData
	}
	return nil
}

type Plugin struct {
	Type  Type   `protobuf:"varint,1,opt,name=Type,proto3,enum=state.Type" json:"Type,omitempty"`
	PData []byte `protobuf:"bytes,2,opt,name=PData,proto3" json:"PData,omitempty"`
}

func (m *Plugin) Reset()                    { *m = Plugin{} }
func (m *Plugin) String() string            { return proto.CompactTextString(m) }
func (*Plugin) ProtoMessage()               {}
func (*Plugin) Descriptor() ([]byte, []int) { return fileDescriptorState, []int{5} }

func (m *Plugin) GetType() Type {
	if m != nil {
		return m.Type
	}
	return Type_PluginNone
}

func (m *Plugin) GetPData() []byte {
	if m != nil {
		return m.PData
	}
	return nil
}

type State struct {
	GenesisDoc         *GenesisDoc    `protobuf:"bytes,1,opt,name=GenesisDoc" json:"GenesisDoc,omitempty"`
	ChainID            string         `protobuf:"bytes,2,opt,name=ChainID,proto3" json:"ChainID,omitempty"`
	LastBlockHeight    int64          `protobuf:"varint,3,opt,name=LastBlockHeight,proto3" json:"LastBlockHeight,omitempty"`
	LastBlockID        *types.BlockID `protobuf:"bytes,4,opt,name=LastBlockID" json:"LastBlockID,omitempty"`
	LastBlockTime      int64          `protobuf:"varint,5,opt,name=LastBlockTime,proto3" json:"LastBlockTime,omitempty"`
	Validators         *ValidatorSet  `protobuf:"bytes,6,opt,name=Validators" json:"Validators,omitempty"`
	LastValidators     *ValidatorSet  `protobuf:"bytes,7,opt,name=LastValidators" json:"LastValidators,omitempty"`
	LastNonEmptyHeight int64          `protobuf:"varint,8,opt,name=LastNonEmptyHeight,proto3" json:"LastNonEmptyHeight,omitempty"`
	AppHash            []byte         `protobuf:"bytes,9,opt,name=AppHash,proto3" json:"AppHash,omitempty"`
	ReceiptsHash       []byte         `protobuf:"bytes,10,opt,name=ReceiptsHash,proto3" json:"ReceiptsHash,omitempty"`
}

func (m *State) Reset()                    { *m = State{} }
func (m *State) String() string            { return proto.CompactTextString(m) }
func (*State) ProtoMessage()               {}
func (*State) Descriptor() ([]byte, []int) { return fileDescriptorState, []int{6} }

func (m *State) GetGenesisDoc() *GenesisDoc {
	if m != nil {
		return m.GenesisDoc
	}
	return nil
}

func (m *State) GetChainID() string {
	if m != nil {
		return m.ChainID
	}
	return ""
}

func (m *State) GetLastBlockHeight() int64 {
	if m != nil {
		return m.LastBlockHeight
	}
	return 0
}

func (m *State) GetLastBlockID() *types.BlockID {
	if m != nil {
		return m.LastBlockID
	}
	return nil
}

func (m *State) GetLastBlockTime() int64 {
	if m != nil {
		return m.LastBlockTime
	}
	return 0
}

func (m *State) GetValidators() *ValidatorSet {
	if m != nil {
		return m.Validators
	}
	return nil
}

func (m *State) GetLastValidators() *ValidatorSet {
	if m != nil {
		return m.LastValidators
	}
	return nil
}

func (m *State) GetLastNonEmptyHeight() int64 {
	if m != nil {
		return m.LastNonEmptyHeight
	}
	return 0
}

func (m *State) GetAppHash() []byte {
	if m != nil {
		return m.AppHash
	}
	return nil
}

func (m *State) GetReceiptsHash() []byte {
	if m != nil {
		return m.ReceiptsHash
	}
	return nil
}

func init() {
	proto.RegisterType((*GenesisDoc)(nil), "state.GenesisDoc")
	proto.RegisterType((*ValidatorSet)(nil), "state.ValidatorSet")
	proto.RegisterType((*SpecialOp)(nil), "state.SpecialOp")
	proto.RegisterType((*SuspectPlugin)(nil), "state.SuspectPlugin")
	proto.RegisterType((*QueryCachePlugin)(nil), "state.QueryCachePlugin")
	proto.RegisterType((*Plugin)(nil), "state.Plugin")
	proto.RegisterType((*State)(nil), "state.State")
	proto.RegisterEnum("state.Type", Type_name, Type_value)
}
func (m *GenesisDoc) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *GenesisDoc) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if len(m.JSONData) > 0 {
		dAtA[i] = 0xa
		i++
		i = encodeVarintState(dAtA, i, uint64(len(m.JSONData)))
		i += copy(dAtA[i:], m.JSONData)
	}
	return i, nil
}

func (m *ValidatorSet) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *ValidatorSet) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if len(m.JSONData) > 0 {
		dAtA[i] = 0x12
		i++
		i = encodeVarintState(dAtA, i, uint64(len(m.JSONData)))
		i += copy(dAtA[i:], m.JSONData)
	}
	return i, nil
}

func (m *SpecialOp) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *SpecialOp) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if len(m.JSONData) > 0 {
		dAtA[i] = 0xa
		i++
		i = encodeVarintState(dAtA, i, uint64(len(m.JSONData)))
		i += copy(dAtA[i:], m.JSONData)
	}
	return i, nil
}

func (m *SuspectPlugin) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *SuspectPlugin) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if len(m.JSONData) > 0 {
		dAtA[i] = 0xa
		i++
		i = encodeVarintState(dAtA, i, uint64(len(m.JSONData)))
		i += copy(dAtA[i:], m.JSONData)
	}
	return i, nil
}

func (m *QueryCachePlugin) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryCachePlugin) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if len(m.JSONData) > 0 {
		dAtA[i] = 0xa
		i++
		i = encodeVarintState(dAtA, i, uint64(len(m.JSONData)))
		i += copy(dAtA[i:], m.JSONData)
	}
	return i, nil
}

func (m *Plugin) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Plugin) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if m.Type != 0 {
		dAtA[i] = 0x8
		i++
		i = encodeVarintState(dAtA, i, uint64(m.Type))
	}
	if len(m.PData) > 0 {
		dAtA[i] = 0x12
		i++
		i = encodeVarintState(dAtA, i, uint64(len(m.PData)))
		i += copy(dAtA[i:], m.PData)
	}
	return i, nil
}

func (m *State) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *State) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if m.GenesisDoc != nil {
		dAtA[i] = 0xa
		i++
		i = encodeVarintState(dAtA, i, uint64(m.GenesisDoc.Size()))
		n1, err := m.GenesisDoc.MarshalTo(dAtA[i:])
		if err != nil {
			return 0, err
		}
		i += n1
	}
	if len(m.ChainID) > 0 {
		dAtA[i] = 0x12
		i++
		i = encodeVarintState(dAtA, i, uint64(len(m.ChainID)))
		i += copy(dAtA[i:], m.ChainID)
	}
	if m.LastBlockHeight != 0 {
		dAtA[i] = 0x18
		i++
		i = encodeVarintState(dAtA, i, uint64(m.LastBlockHeight))
	}
	if m.LastBlockID != nil {
		dAtA[i] = 0x22
		i++
		i = encodeVarintState(dAtA, i, uint64(m.LastBlockID.Size()))
		n2, err := m.LastBlockID.MarshalTo(dAtA[i:])
		if err != nil {
			return 0, err
		}
		i += n2
	}
	if m.LastBlockTime != 0 {
		dAtA[i] = 0x28
		i++
		i = encodeVarintState(dAtA, i, uint64(m.LastBlockTime))
	}
	if m.Validators != nil {
		dAtA[i] = 0x32
		i++
		i = encodeVarintState(dAtA, i, uint64(m.Validators.Size()))
		n3, err := m.Validators.MarshalTo(dAtA[i:])
		if err != nil {
			return 0, err
		}
		i += n3
	}
	if m.LastValidators != nil {
		dAtA[i] = 0x3a
		i++
		i = encodeVarintState(dAtA, i, uint64(m.LastValidators.Size()))
		n4, err := m.LastValidators.MarshalTo(dAtA[i:])
		if err != nil {
			return 0, err
		}
		i += n4
	}
	if m.LastNonEmptyHeight != 0 {
		dAtA[i] = 0x40
		i++
		i = encodeVarintState(dAtA, i, uint64(m.LastNonEmptyHeight))
	}
	if len(m.AppHash) > 0 {
		dAtA[i] = 0x4a
		i++
		i = encodeVarintState(dAtA, i, uint64(len(m.AppHash)))
		i += copy(dAtA[i:], m.AppHash)
	}
	if len(m.ReceiptsHash) > 0 {
		dAtA[i] = 0x52
		i++
		i = encodeVarintState(dAtA, i, uint64(len(m.ReceiptsHash)))
		i += copy(dAtA[i:], m.ReceiptsHash)
	}
	return i, nil
}

func encodeFixed64State(dAtA []byte, offset int, v uint64) int {
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
func encodeFixed32State(dAtA []byte, offset int, v uint32) int {
	dAtA[offset] = uint8(v)
	dAtA[offset+1] = uint8(v >> 8)
	dAtA[offset+2] = uint8(v >> 16)
	dAtA[offset+3] = uint8(v >> 24)
	return offset + 4
}
func encodeVarintState(dAtA []byte, offset int, v uint64) int {
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return offset + 1
}
func (m *GenesisDoc) Size() (n int) {
	var l int
	_ = l
	l = len(m.JSONData)
	if l > 0 {
		n += 1 + l + sovState(uint64(l))
	}
	return n
}

func (m *ValidatorSet) Size() (n int) {
	var l int
	_ = l
	l = len(m.JSONData)
	if l > 0 {
		n += 1 + l + sovState(uint64(l))
	}
	return n
}

func (m *SpecialOp) Size() (n int) {
	var l int
	_ = l
	l = len(m.JSONData)
	if l > 0 {
		n += 1 + l + sovState(uint64(l))
	}
	return n
}

func (m *SuspectPlugin) Size() (n int) {
	var l int
	_ = l
	l = len(m.JSONData)
	if l > 0 {
		n += 1 + l + sovState(uint64(l))
	}
	return n
}

func (m *QueryCachePlugin) Size() (n int) {
	var l int
	_ = l
	l = len(m.JSONData)
	if l > 0 {
		n += 1 + l + sovState(uint64(l))
	}
	return n
}

func (m *Plugin) Size() (n int) {
	var l int
	_ = l
	if m.Type != 0 {
		n += 1 + sovState(uint64(m.Type))
	}
	l = len(m.PData)
	if l > 0 {
		n += 1 + l + sovState(uint64(l))
	}
	return n
}

func (m *State) Size() (n int) {
	var l int
	_ = l
	if m.GenesisDoc != nil {
		l = m.GenesisDoc.Size()
		n += 1 + l + sovState(uint64(l))
	}
	l = len(m.ChainID)
	if l > 0 {
		n += 1 + l + sovState(uint64(l))
	}
	if m.LastBlockHeight != 0 {
		n += 1 + sovState(uint64(m.LastBlockHeight))
	}
	if m.LastBlockID != nil {
		l = m.LastBlockID.Size()
		n += 1 + l + sovState(uint64(l))
	}
	if m.LastBlockTime != 0 {
		n += 1 + sovState(uint64(m.LastBlockTime))
	}
	if m.Validators != nil {
		l = m.Validators.Size()
		n += 1 + l + sovState(uint64(l))
	}
	if m.LastValidators != nil {
		l = m.LastValidators.Size()
		n += 1 + l + sovState(uint64(l))
	}
	if m.LastNonEmptyHeight != 0 {
		n += 1 + sovState(uint64(m.LastNonEmptyHeight))
	}
	l = len(m.AppHash)
	if l > 0 {
		n += 1 + l + sovState(uint64(l))
	}
	l = len(m.ReceiptsHash)
	if l > 0 {
		n += 1 + l + sovState(uint64(l))
	}
	return n
}

func sovState(x uint64) (n int) {
	for {
		n++
		x >>= 7
		if x == 0 {
			break
		}
	}
	return n
}
func sozState(x uint64) (n int) {
	return sovState(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *GenesisDoc) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowState
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
			return fmt.Errorf("proto: GenesisDoc: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: GenesisDoc: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field JSONData", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowState
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
				return ErrInvalidLengthState
			}
			postIndex := iNdEx + byteLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.JSONData = append(m.JSONData[:0], dAtA[iNdEx:postIndex]...)
			if m.JSONData == nil {
				m.JSONData = []byte{}
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipState(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthState
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
func (m *ValidatorSet) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowState
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
			return fmt.Errorf("proto: ValidatorSet: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: ValidatorSet: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field JSONData", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowState
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
				return ErrInvalidLengthState
			}
			postIndex := iNdEx + byteLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.JSONData = append(m.JSONData[:0], dAtA[iNdEx:postIndex]...)
			if m.JSONData == nil {
				m.JSONData = []byte{}
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipState(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthState
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
func (m *SpecialOp) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowState
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
			return fmt.Errorf("proto: SpecialOp: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: SpecialOp: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field JSONData", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowState
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
				return ErrInvalidLengthState
			}
			postIndex := iNdEx + byteLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.JSONData = append(m.JSONData[:0], dAtA[iNdEx:postIndex]...)
			if m.JSONData == nil {
				m.JSONData = []byte{}
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipState(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthState
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
func (m *SuspectPlugin) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowState
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
			return fmt.Errorf("proto: SuspectPlugin: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: SuspectPlugin: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field JSONData", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowState
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
				return ErrInvalidLengthState
			}
			postIndex := iNdEx + byteLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.JSONData = append(m.JSONData[:0], dAtA[iNdEx:postIndex]...)
			if m.JSONData == nil {
				m.JSONData = []byte{}
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipState(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthState
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
func (m *QueryCachePlugin) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowState
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
			return fmt.Errorf("proto: QueryCachePlugin: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryCachePlugin: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field JSONData", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowState
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
				return ErrInvalidLengthState
			}
			postIndex := iNdEx + byteLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.JSONData = append(m.JSONData[:0], dAtA[iNdEx:postIndex]...)
			if m.JSONData == nil {
				m.JSONData = []byte{}
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipState(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthState
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
func (m *Plugin) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowState
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
			return fmt.Errorf("proto: Plugin: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Plugin: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Type", wireType)
			}
			m.Type = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowState
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Type |= (Type(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field PData", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowState
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
				return ErrInvalidLengthState
			}
			postIndex := iNdEx + byteLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.PData = append(m.PData[:0], dAtA[iNdEx:postIndex]...)
			if m.PData == nil {
				m.PData = []byte{}
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipState(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthState
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
func (m *State) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowState
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
			return fmt.Errorf("proto: State: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: State: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field GenesisDoc", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowState
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
				return ErrInvalidLengthState
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.GenesisDoc == nil {
				m.GenesisDoc = &GenesisDoc{}
			}
			if err := m.GenesisDoc.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ChainID", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowState
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthState
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.ChainID = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field LastBlockHeight", wireType)
			}
			m.LastBlockHeight = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowState
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.LastBlockHeight |= (int64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field LastBlockID", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowState
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
				return ErrInvalidLengthState
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.LastBlockID == nil {
				m.LastBlockID = &types.BlockID{}
			}
			if err := m.LastBlockID.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 5:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field LastBlockTime", wireType)
			}
			m.LastBlockTime = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowState
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.LastBlockTime |= (int64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 6:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Validators", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowState
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
				return ErrInvalidLengthState
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.Validators == nil {
				m.Validators = &ValidatorSet{}
			}
			if err := m.Validators.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 7:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field LastValidators", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowState
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
				return ErrInvalidLengthState
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.LastValidators == nil {
				m.LastValidators = &ValidatorSet{}
			}
			if err := m.LastValidators.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 8:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field LastNonEmptyHeight", wireType)
			}
			m.LastNonEmptyHeight = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowState
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.LastNonEmptyHeight |= (int64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 9:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field AppHash", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowState
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
				return ErrInvalidLengthState
			}
			postIndex := iNdEx + byteLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.AppHash = append(m.AppHash[:0], dAtA[iNdEx:postIndex]...)
			if m.AppHash == nil {
				m.AppHash = []byte{}
			}
			iNdEx = postIndex
		case 10:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ReceiptsHash", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowState
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
				return ErrInvalidLengthState
			}
			postIndex := iNdEx + byteLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.ReceiptsHash = append(m.ReceiptsHash[:0], dAtA[iNdEx:postIndex]...)
			if m.ReceiptsHash == nil {
				m.ReceiptsHash = []byte{}
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipState(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthState
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
func skipState(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowState
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
					return 0, ErrIntOverflowState
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
					return 0, ErrIntOverflowState
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
				return 0, ErrInvalidLengthState
			}
			return iNdEx, nil
		case 3:
			for {
				var innerWire uint64
				var start int = iNdEx
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return 0, ErrIntOverflowState
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
				next, err := skipState(dAtA[start:])
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
	ErrInvalidLengthState = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowState   = fmt.Errorf("proto: integer overflow")
)

func init() { proto.RegisterFile("state.proto", fileDescriptorState) }

var fileDescriptorState = []byte{
	// 470 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x8c, 0x53, 0xcd, 0x6e, 0xd3, 0x40,
	0x10, 0xae, 0x9b, 0x26, 0x6d, 0x26, 0x3f, 0x75, 0xb7, 0x3d, 0x58, 0x3d, 0x84, 0x2a, 0x42, 0x22,
	0x2a, 0x92, 0x03, 0xed, 0x11, 0x24, 0x44, 0x1b, 0x44, 0x8b, 0x50, 0x5a, 0xd6, 0x11, 0xf7, 0xad,
	0x19, 0xd9, 0x2b, 0x9c, 0xdd, 0x55, 0x76, 0x73, 0x08, 0x4f, 0xc2, 0x93, 0xf0, 0x0c, 0x1c, 0x79,
	0x04, 0x14, 0x5e, 0x04, 0x79, 0xd7, 0x75, 0x9d, 0x08, 0xa2, 0x5e, 0x2c, 0x7d, 0xdf, 0x7c, 0xf3,
	0xd9, 0x33, 0xdf, 0x18, 0x5a, 0xda, 0x30, 0x83, 0xa1, 0x9a, 0x49, 0x23, 0x49, 0xdd, 0x82, 0xe3,
	0xd7, 0x09, 0x37, 0x19, 0xbb, 0x0b, 0xbf, 0xa5, 0x52, 0x24, 0x4c, 0x48, 0x91, 0x71, 0x81, 0x61,
	0x2c, 0xa7, 0x43, 0x26, 0xc4, 0x90, 0x89, 0x84, 0x0b, 0x1c, 0xda, 0x0e, 0x3d, 0x34, 0x0b, 0x85,
	0xc5, 0xd3, 0x99, 0xf4, 0x07, 0x00, 0xef, 0x51, 0xa0, 0xe6, 0x7a, 0x24, 0x63, 0x72, 0x0c, 0x7b,
	0x1f, 0xa2, 0x9b, 0xf1, 0x88, 0x19, 0x16, 0x78, 0x27, 0xde, 0xa0, 0x4d, 0x4b, 0xdc, 0x3f, 0x85,
	0xf6, 0x67, 0x96, 0xf1, 0x2f, 0xcc, 0xc8, 0x59, 0x84, 0x66, 0x45, 0xbb, 0xbd, 0xa6, 0x7d, 0x06,
	0xcd, 0x48, 0x61, 0xcc, 0x59, 0x76, 0xa3, 0x36, 0x9a, 0x3e, 0x87, 0x4e, 0x34, 0xd7, 0x0a, 0x63,
	0x73, 0x9b, 0xcd, 0x13, 0x2e, 0x36, 0x8a, 0x43, 0xf0, 0x3f, 0xcd, 0x71, 0xb6, 0xb8, 0x64, 0x71,
	0x8a, 0x8f, 0xd0, 0xbf, 0x81, 0x46, 0xa1, 0x7a, 0x02, 0x3b, 0x93, 0x85, 0x42, 0xab, 0xe8, 0x9e,
	0xb5, 0x42, 0xb7, 0xc6, 0x9c, 0xa2, 0xb6, 0x40, 0x8e, 0xa0, 0x7e, 0x5b, 0x99, 0xc4, 0x81, 0xfe,
	0x8f, 0x1a, 0xd4, 0xa3, 0x5c, 0x4a, 0x5e, 0x56, 0xd7, 0x64, 0x6d, 0x5a, 0x67, 0x07, 0x85, 0xcd,
	0x43, 0x81, 0x56, 0x77, 0x19, 0xc0, 0xee, 0x65, 0xca, 0xb8, 0xb8, 0x1e, 0x59, 0xd3, 0x26, 0xbd,
	0x87, 0x64, 0x00, 0xfb, 0x1f, 0x99, 0x36, 0x17, 0x99, 0x8c, 0xbf, 0x5e, 0x21, 0x4f, 0x52, 0x13,
	0xd4, 0x4e, 0xbc, 0x41, 0x8d, 0xae, 0xd3, 0xe4, 0x05, 0xb4, 0x4a, 0xea, 0x7a, 0x14, 0xec, 0xd8,
	0xf7, 0x76, 0x43, 0x17, 0x60, 0xc1, 0xd2, 0xaa, 0x84, 0x3c, 0x85, 0x4e, 0x09, 0x27, 0x7c, 0x8a,
	0x41, 0xdd, 0x3a, 0xaf, 0x92, 0xe4, 0x1c, 0xa0, 0xcc, 0x52, 0x07, 0x0d, 0x6b, 0x7b, 0x58, 0x8c,
	0x53, 0x0d, 0x99, 0x56, 0x64, 0xe4, 0x15, 0x74, 0x73, 0x97, 0x4a, 0xe3, 0xee, 0xff, 0x1b, 0xd7,
	0xa4, 0x24, 0x04, 0x92, 0x33, 0x63, 0x29, 0xde, 0x4d, 0x95, 0x59, 0x14, 0x63, 0xef, 0xd9, 0x8f,
	0xfb, 0x47, 0x25, 0xdf, 0xde, 0x5b, 0xa5, 0xae, 0x98, 0x4e, 0x83, 0xa6, 0x8d, 0xe4, 0x1e, 0x92,
	0x3e, 0xb4, 0x29, 0xc6, 0xc8, 0x95, 0xd1, 0xb6, 0x0c, 0xb6, 0xbc, 0xc2, 0x9d, 0x4e, 0x5c, 0xde,
	0xa4, 0x0b, 0xe0, 0x2e, 0x60, 0x2c, 0x05, 0xfa, 0x5b, 0xe4, 0x10, 0xf6, 0x1d, 0x2e, 0xaf, 0xd3,
	0xf7, 0xc8, 0x01, 0x74, 0x0a, 0xd2, 0x5d, 0xa2, 0xbf, 0x4d, 0x8e, 0xc0, 0x77, 0xd4, 0xc3, 0xbd,
	0xf9, 0xb5, 0x0b, 0xff, 0xe7, 0xb2, 0xe7, 0xfd, 0x5a, 0xf6, 0xbc, 0xdf, 0xcb, 0x9e, 0xf7, 0xfd,
	0x4f, 0x6f, 0xeb, 0xae, 0x61, 0x7f, 0xa2, 0xf3, 0xbf, 0x01, 0x00, 0x00, 0xff, 0xff, 0x93, 0x7a,
	0x34, 0xc4, 0x98, 0x03, 0x00, 0x00,
}
