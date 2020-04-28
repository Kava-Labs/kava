package types

import (
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
)

type InterA interface {
	GetTitle() string
}

type InterB InterA

// interface {
// 	GetDescription() string
// }

type Prop1 struct{}

func (p Prop1) GetTitle() string       { return "prop1 title" }
func (p Prop1) GetDescription() string { return "prop1 description" }

type Prop2 struct{}

func (p Prop2) GetTitle() string       { return "prop2 title" }
func (p Prop2) GetDescription() string { return "prop2 description" }

func TestTest(t *testing.T) {
	/*
		register content, register new pubproposal
		register concrete types (should satisfy both of them)

		try marshalling and unmarshalling all 4 combinations
	*/
	cdc := codec.New()

	cdc.RegisterInterface((*InterA)(nil), nil)
	cdc.RegisterConcrete(Prop1{}, "test/prop1", nil)
	cdc.RegisterInterface((*InterB)(nil), nil)
	cdc.RegisterConcrete(Prop2{}, "test/prop2", nil)

	p1ia := InterA(Prop1{})
	p2ia := InterA(Prop2{})
	p1ib := InterB(Prop1{})
	p2ib := InterB(Prop2{})

	var iap1 InterA
	cdc.MustUnmarshalBinaryBare(cdc.MustMarshalBinaryBare(p1ia), &iap1)
	fmt.Printf("%T, %T\n", p1ia, iap1)

	var iap2 InterA
	cdc.MustUnmarshalBinaryBare(cdc.MustMarshalBinaryBare(p2ia), &iap2)
	fmt.Printf("%T, %T\n", p2ia, iap2)

	var ibp1 InterB
	cdc.MustUnmarshalBinaryBare(cdc.MustMarshalBinaryBare(p1ib), &ibp1)
	fmt.Printf("%T, %T\n", p1ib, ibp1)

	var ibp2 InterB
	cdc.MustUnmarshalBinaryBare(cdc.MustMarshalBinaryBare(p2ib), &ibp2)
	fmt.Printf("%T, %T\n", p2ib, ibp2)
}
