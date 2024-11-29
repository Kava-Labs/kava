// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: kava/precisebank/v1/genesis.proto

package types

import (
	cosmossdk_io_math "cosmossdk.io/math"
	fmt "fmt"
	_ "github.com/cosmos/cosmos-proto"
	_ "github.com/cosmos/gogoproto/gogoproto"
	proto "github.com/cosmos/gogoproto/proto"
	io "io"
	math "math"
	math_bits "math/bits"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion3 // please upgrade the proto package

// GenesisState defines the precisebank module's genesis state.
type GenesisState struct {
	// balances is a list of all the balances in the precisebank module.
	Balances FractionalBalances `protobuf:"bytes,1,rep,name=balances,proto3,castrepeated=FractionalBalances" json:"balances"`
	// remainder is an internal value of how much extra fractional digits are
	// still backed by the reserve, but not assigned to any account.
	Remainder cosmossdk_io_math.Int `protobuf:"bytes,2,opt,name=remainder,proto3,customtype=cosmossdk.io/math.Int" json:"remainder"`
}

func (m *GenesisState) Reset()         { *m = GenesisState{} }
func (m *GenesisState) String() string { return proto.CompactTextString(m) }
func (*GenesisState) ProtoMessage()    {}
func (*GenesisState) Descriptor() ([]byte, []int) {
	return fileDescriptor_7f1c47a86fb0d2e0, []int{0}
}
func (m *GenesisState) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *GenesisState) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_GenesisState.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *GenesisState) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GenesisState.Merge(m, src)
}
func (m *GenesisState) XXX_Size() int {
	return m.Size()
}
func (m *GenesisState) XXX_DiscardUnknown() {
	xxx_messageInfo_GenesisState.DiscardUnknown(m)
}

var xxx_messageInfo_GenesisState proto.InternalMessageInfo

func (m *GenesisState) GetBalances() FractionalBalances {
	if m != nil {
		return m.Balances
	}
	return nil
}

// FractionalBalance defines the fractional portion of an account balance
type FractionalBalance struct {
	// address is the address of the balance holder.
	Address string `protobuf:"bytes,1,opt,name=address,proto3" json:"address,omitempty"`
	// amount indicates amount of only the fractional balance owned by the
	// address. FractionalBalance currently only supports tracking 1 single asset,
	// e.g. fractional balances of ukava.
	Amount cosmossdk_io_math.Int `protobuf:"bytes,2,opt,name=amount,proto3,customtype=cosmossdk.io/math.Int" json:"amount"`
}

func (m *FractionalBalance) Reset()         { *m = FractionalBalance{} }
func (m *FractionalBalance) String() string { return proto.CompactTextString(m) }
func (*FractionalBalance) ProtoMessage()    {}
func (*FractionalBalance) Descriptor() ([]byte, []int) {
	return fileDescriptor_7f1c47a86fb0d2e0, []int{1}
}
func (m *FractionalBalance) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *FractionalBalance) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_FractionalBalance.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *FractionalBalance) XXX_Merge(src proto.Message) {
	xxx_messageInfo_FractionalBalance.Merge(m, src)
}
func (m *FractionalBalance) XXX_Size() int {
	return m.Size()
}
func (m *FractionalBalance) XXX_DiscardUnknown() {
	xxx_messageInfo_FractionalBalance.DiscardUnknown(m)
}

var xxx_messageInfo_FractionalBalance proto.InternalMessageInfo

func init() {
	proto.RegisterType((*GenesisState)(nil), "kava.precisebank.v1.GenesisState")
	proto.RegisterType((*FractionalBalance)(nil), "kava.precisebank.v1.FractionalBalance")
}

func init() { proto.RegisterFile("kava/precisebank/v1/genesis.proto", fileDescriptor_7f1c47a86fb0d2e0) }

var fileDescriptor_7f1c47a86fb0d2e0 = []byte{
	// 351 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x9c, 0x91, 0x3f, 0x4f, 0xc2, 0x40,
	0x18, 0xc6, 0x7b, 0x9a, 0x20, 0x9c, 0x2e, 0x56, 0x4c, 0x90, 0xa1, 0x45, 0x06, 0x43, 0x62, 0x7a,
	0x17, 0x70, 0x73, 0xb3, 0x26, 0x12, 0x56, 0xd8, 0x1c, 0x34, 0xd7, 0xf6, 0x52, 0x2e, 0xd0, 0x3b,
	0x72, 0x77, 0x10, 0xfd, 0x06, 0x8e, 0x4e, 0xce, 0xcc, 0xce, 0x2c, 0x7e, 0x03, 0x46, 0xc2, 0x64,
	0x1c, 0xd0, 0xc0, 0xe2, 0xc7, 0x30, 0xe5, 0xf0, 0x5f, 0x70, 0x72, 0x7b, 0xfb, 0xbe, 0xbf, 0xe7,
	0x79, 0x9f, 0xde, 0x0b, 0x0f, 0x3b, 0x64, 0x40, 0x70, 0x4f, 0xd2, 0x90, 0x29, 0x1a, 0x10, 0xde,
	0xc1, 0x83, 0x2a, 0x8e, 0x29, 0xa7, 0x8a, 0x29, 0xd4, 0x93, 0x42, 0x0b, 0x7b, 0x2f, 0x45, 0xd0,
	0x0f, 0x04, 0x0d, 0xaa, 0xc5, 0x83, 0x50, 0xa8, 0x44, 0xa8, 0xeb, 0x25, 0x82, 0xcd, 0x87, 0xe1,
	0x8b, 0xf9, 0x58, 0xc4, 0xc2, 0xf4, 0xd3, 0xca, 0x74, 0xcb, 0x4f, 0x00, 0xee, 0xd4, 0x8d, 0x6f,
	0x4b, 0x13, 0x4d, 0xed, 0x2b, 0x98, 0x0d, 0x48, 0x97, 0xf0, 0x90, 0xaa, 0x02, 0x28, 0x6d, 0x56,
	0xb6, 0x6b, 0x47, 0xe8, 0x8f, 0x4d, 0xe8, 0x42, 0x92, 0x50, 0x33, 0xc1, 0x49, 0xd7, 0x37, 0xb8,
	0x5f, 0x1c, 0xcf, 0x5c, 0xeb, 0xf1, 0xd5, 0xb5, 0xd7, 0x46, 0xaa, 0xf9, 0xe5, 0x69, 0x37, 0x60,
	0x4e, 0xd2, 0x84, 0x30, 0x1e, 0x51, 0x59, 0xd8, 0x28, 0x81, 0x4a, 0xce, 0x3f, 0x4e, 0x85, 0x2f,
	0x33, 0x77, 0xdf, 0xe4, 0x55, 0x51, 0x07, 0x31, 0x81, 0x13, 0xa2, 0xdb, 0xa8, 0xc1, 0xf5, 0x74,
	0xe4, 0xc1, 0xd5, 0x8f, 0x34, 0xb8, 0x6e, 0x7e, 0xab, 0xcb, 0x0f, 0x00, 0xee, 0xae, 0xed, 0xb2,
	0x6b, 0x70, 0x8b, 0x44, 0x91, 0xa4, 0x2a, 0xcd, 0x9f, 0xda, 0x17, 0xa6, 0x23, 0x2f, 0xbf, 0x72,
	0x38, 0x33, 0x93, 0x96, 0x96, 0x8c, 0xc7, 0xcd, 0x4f, 0xd0, 0x3e, 0x87, 0x19, 0x92, 0x88, 0x3e,
	0xd7, 0xff, 0x49, 0xb4, 0x92, 0x9e, 0x66, 0xef, 0x86, 0xae, 0xf5, 0x3e, 0x74, 0x2d, 0xbf, 0x3e,
	0x9e, 0x3b, 0x60, 0x32, 0x77, 0xc0, 0xdb, 0xdc, 0x01, 0xf7, 0x0b, 0xc7, 0x9a, 0x2c, 0x1c, 0xeb,
	0x79, 0xe1, 0x58, 0x97, 0x5e, 0xcc, 0x74, 0xbb, 0x1f, 0xa0, 0x50, 0x24, 0x38, 0x7d, 0x55, 0xaf,
	0x4b, 0x02, 0xb5, 0xac, 0xf0, 0xcd, 0xaf, 0x73, 0xeb, 0xdb, 0x1e, 0x55, 0x41, 0x66, 0x79, 0xa4,
	0x93, 0x8f, 0x00, 0x00, 0x00, 0xff, 0xff, 0xbd, 0xf6, 0xf2, 0x8d, 0x0f, 0x02, 0x00, 0x00,
}

func (m *GenesisState) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *GenesisState) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *GenesisState) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	{
		size := m.Remainder.Size()
		i -= size
		if _, err := m.Remainder.MarshalTo(dAtA[i:]); err != nil {
			return 0, err
		}
		i = encodeVarintGenesis(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x12
	if len(m.Balances) > 0 {
		for iNdEx := len(m.Balances) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.Balances[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintGenesis(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0xa
		}
	}
	return len(dAtA) - i, nil
}

func (m *FractionalBalance) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *FractionalBalance) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *FractionalBalance) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	{
		size := m.Amount.Size()
		i -= size
		if _, err := m.Amount.MarshalTo(dAtA[i:]); err != nil {
			return 0, err
		}
		i = encodeVarintGenesis(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x12
	if len(m.Address) > 0 {
		i -= len(m.Address)
		copy(dAtA[i:], m.Address)
		i = encodeVarintGenesis(dAtA, i, uint64(len(m.Address)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func encodeVarintGenesis(dAtA []byte, offset int, v uint64) int {
	offset -= sovGenesis(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *GenesisState) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if len(m.Balances) > 0 {
		for _, e := range m.Balances {
			l = e.Size()
			n += 1 + l + sovGenesis(uint64(l))
		}
	}
	l = m.Remainder.Size()
	n += 1 + l + sovGenesis(uint64(l))
	return n
}

func (m *FractionalBalance) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Address)
	if l > 0 {
		n += 1 + l + sovGenesis(uint64(l))
	}
	l = m.Amount.Size()
	n += 1 + l + sovGenesis(uint64(l))
	return n
}

func sovGenesis(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozGenesis(x uint64) (n int) {
	return sovGenesis(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *GenesisState) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowGenesis
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: GenesisState: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: GenesisState: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Balances", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Balances = append(m.Balances, FractionalBalance{})
			if err := m.Balances[len(m.Balances)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Remainder", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.Remainder.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipGenesis(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthGenesis
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
func (m *FractionalBalance) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowGenesis
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: FractionalBalance: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: FractionalBalance: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Address", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Address = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Amount", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.Amount.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipGenesis(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthGenesis
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
func skipGenesis(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowGenesis
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
					return 0, ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
					break
				}
			}
		case 1:
			iNdEx += 8
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowGenesis
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
			if length < 0 {
				return 0, ErrInvalidLengthGenesis
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupGenesis
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthGenesis
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthGenesis        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowGenesis          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupGenesis = fmt.Errorf("proto: unexpected end of group")
)