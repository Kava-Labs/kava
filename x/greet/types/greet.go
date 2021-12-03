package types

import (
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)


const (
	ModuleName = "greet"
	StoreKey = ModuleName
	RouterKey = ModuleName
	QuerierRoute = ModuleName
	GreetKey = "greet-value-"
	GreetCountKey = "greet-count-"
	QueryGetGreeting = "get-greeting"
	QueryListGreetings = "list-greetings"
)

func KeyPrefix(p string) []byte {
	return []byte(p)
}

func DefaultGenesisState() GenesisState {
	return GenesisState{}
}

func (gs GenesisState) Validate() error {
	return nil
}


var _ sdk.Msg = &MsgCreateGreet{}


func NewMsgCreateGreet(owner string, message string) *MsgCreateGreet{
	return &MsgCreateGreet{
		Owner: owner,
		Message: message,
	}
}

func (m *MsgCreateGreet) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Owner); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid owner address", err)
	}
	
	if len(strings.TrimSpace(m.Message)) == 0 {
		return fmt.Errorf("must provide a greeting message")
	}

	return nil
}

func (m *MsgCreateGreet) GetSigners() []sdk.AccAddress {
	owner, err := sdk.AccAddressFromBech32(m.Owner); 
	if err != nil {
		panic(err)
	}
	
	return []sdk.AccAddress{owner}
}



func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino){
	cdc.RegisterConcrete(&MsgCreateGreet{}, "greet/CreateGreet", nil)
}

func RegisterInterfaces(registry types.InterfaceRegistry){
	registry.RegisterImplementations((*sdk.Msg)(nil), &MsgCreateGreet{})
	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}



var amino = codec.NewLegacyAmino()
var ModuleCdc = codec.NewAminoCodec(amino)
