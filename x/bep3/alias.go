package bep3

import (
	"github.com/kava-labs/kava/x/bep3/client/rest"
	"github.com/kava-labs/kava/x/bep3/keeper"
	"github.com/kava-labs/kava/x/bep3/types"
)

const (
	AddrByteCount                  = types.AddrByteCount
	AttributeKeyAmount             = types.AttributeKeyAmount
	AttributeKeyAtomicSwapID       = types.AttributeKeyAtomicSwapID
	AttributeKeyAtomicSwapIDs      = types.AttributeKeyAtomicSwapIDs
	AttributeKeyClaimSender        = types.AttributeKeyClaimSender
	AttributeKeyDirection          = types.AttributeKeyDirection
	AttributeKeyExpectedIncome     = types.AttributeKeyExpectedIncome
	AttributeKeyExpireHeight       = types.AttributeKeyExpireHeight
	AttributeKeyRandomNumber       = types.AttributeKeyRandomNumber
	AttributeKeyRandomNumberHash   = types.AttributeKeyRandomNumberHash
	AttributeKeyRecipient          = types.AttributeKeyRecipient
	AttributeKeyRefundSender       = types.AttributeKeyRefundSender
	AttributeKeySender             = types.AttributeKeySender
	AttributeKeySenderOtherChain   = types.AttributeKeySenderOtherChain
	AttributeKeyTimestamp          = types.AttributeKeyTimestamp
	AttributeValueCategory         = types.AttributeValueCategory
	CalcSwapID                     = types.CalcSwapID
	ClaimAtomicSwap                = types.ClaimAtomicSwap
	Completed                      = types.Completed
	CreateAtomicSwap               = types.CreateAtomicSwap
	DefaultLongtermStorageDuration = types.DefaultLongtermStorageDuration
	DefaultParamspace              = types.DefaultParamspace
	DepositAtomicSwap              = types.DepositAtomicSwap
	EventTypeClaimAtomicSwap       = types.EventTypeClaimAtomicSwap
	EventTypeCreateAtomicSwap      = types.EventTypeCreateAtomicSwap
	EventTypeRefundAtomicSwap      = types.EventTypeRefundAtomicSwap
	EventTypeSwapsExpired          = types.EventTypeSwapsExpired
	Expired                        = types.Expired
	INVALID                        = types.INVALID
	Incoming                       = types.Incoming
	Int64Size                      = types.Int64Size
	MaxExpectedIncomeLength        = types.MaxExpectedIncomeLength
	MaxOtherChainAddrLength        = types.MaxOtherChainAddrLength
	ModuleName                     = types.ModuleName
	NULL                           = types.NULL
	Open                           = types.Open
	Outgoing                       = types.Outgoing
	QuerierRoute                   = types.QuerierRoute
	QueryGetAssetSupply            = types.QueryGetAssetSupply
	QueryGetAtomicSwap             = types.QueryGetAtomicSwap
	QueryGetAtomicSwaps            = types.QueryGetAtomicSwaps
	QueryGetParams                 = types.QueryGetParams
	RandomNumberHashLength         = types.RandomNumberHashLength
	RandomNumberLength             = types.RandomNumberLength
	RefundAtomicSwap               = types.RefundAtomicSwap
	RouterKey                      = types.RouterKey
	StoreKey                       = types.StoreKey
	SwapIDLength                   = types.SwapIDLength
)

var (
	NewKeeper                  = keeper.NewKeeper
	NewQuerier                 = keeper.NewQuerier
	RegisterRoutes             = rest.RegisterRoutes
	BytesToHex                 = types.BytesToHex
	CalculateRandomHash        = types.CalculateRandomHash
	CalculateSwapID            = types.CalculateSwapID
	DefaultGenesisState        = types.DefaultGenesisState
	DefaultParams              = types.DefaultParams
	ErrAssetNotActive          = types.ErrAssetNotActive
	ErrAssetNotSupported       = types.ErrAssetNotSupported
	ErrAssetSupplyNotFound     = types.ErrAssetSupplyNotFound
	ErrAtomicSwapAlreadyExists = types.ErrAtomicSwapAlreadyExists
	ErrAtomicSwapNotFound      = types.ErrAtomicSwapNotFound
	ErrExceedsAvailableSupply  = types.ErrExceedsAvailableSupply
	ErrExceedsSupplyLimit      = types.ErrExceedsSupplyLimit
	ErrInvalidClaimSecret      = types.ErrInvalidClaimSecret
	ErrInvalidCurrentSupply    = types.ErrInvalidCurrentSupply
	ErrInvalidHeightSpan       = types.ErrInvalidHeightSpan
	ErrInvalidIncomingSupply   = types.ErrInvalidIncomingSupply
	ErrInvalidOutgoingSupply   = types.ErrInvalidOutgoingSupply
	ErrInvalidTimestamp        = types.ErrInvalidTimestamp
	ErrSwapNotClaimable        = types.ErrSwapNotClaimable
	ErrSwapNotRefundable       = types.ErrSwapNotRefundable
	GenerateSecureRandomNumber = types.GenerateSecureRandomNumber
	GetAtomicSwapByHeightKey   = types.GetAtomicSwapByHeightKey
	HexToBytes                 = types.HexToBytes
	NewAssetSupply             = types.NewAssetSupply
	NewAtomicSwap              = types.NewAtomicSwap
	NewGenesisState            = types.NewGenesisState
	NewMsgClaimAtomicSwap      = types.NewMsgClaimAtomicSwap
	NewMsgCreateAtomicSwap     = types.NewMsgCreateAtomicSwap
	NewMsgRefundAtomicSwap     = types.NewMsgRefundAtomicSwap
	NewParams                  = types.NewParams
	NewQueryAssetSupply        = types.NewQueryAssetSupply
	NewQueryAtomicSwapByID     = types.NewQueryAtomicSwapByID
	NewQueryAtomicSwaps        = types.NewQueryAtomicSwaps
	NewSwapDirectionFromString = types.NewSwapDirectionFromString
	NewSwapStatusFromString    = types.NewSwapStatusFromString
	ParamKeyTable              = types.ParamKeyTable
	RegisterCodec              = types.RegisterCodec
	Uint64FromBytes            = types.Uint64FromBytes
	Uint64ToBytes              = types.Uint64ToBytes

	// variable aliases
	AbsoluteMaximumBlockLock        = types.AbsoluteMaximumBlockLock
	AbsoluteMinimumBlockLock        = types.AbsoluteMinimumBlockLock
	AssetSupplyKeyPrefix            = types.AssetSupplyKeyPrefix
	AtomicSwapByBlockPrefix         = types.AtomicSwapByBlockPrefix
	AtomicSwapCoinsAccAddr          = types.AtomicSwapCoinsAccAddr
	AtomicSwapKeyPrefix             = types.AtomicSwapKeyPrefix
	AtomicSwapLongtermStoragePrefix = types.AtomicSwapLongtermStoragePrefix
	DefaultMaxBlockLock             = types.DefaultMaxBlockLock
	DefaultMinBlockLock             = types.DefaultMinBlockLock
	DefaultSupportedAssets          = types.DefaultSupportedAssets
	KeyBnbDeputyAddress             = types.KeyBnbDeputyAddress
	KeyMaxBlockLock                 = types.KeyMaxBlockLock
	KeyMinBlockLock                 = types.KeyMinBlockLock
	KeySupportedAssets              = types.KeySupportedAssets
	ModuleCdc                       = types.ModuleCdc
)

type (
	Keeper              = keeper.Keeper
	AssetParam          = types.AssetParam
	AssetParams         = types.AssetParams
	AssetSupplies       = types.AssetSupplies
	AssetSupply         = types.AssetSupply
	AtomicSwap          = types.AtomicSwap
	AtomicSwaps         = types.AtomicSwaps
	GenesisState        = types.GenesisState
	MsgClaimAtomicSwap  = types.MsgClaimAtomicSwap
	MsgCreateAtomicSwap = types.MsgCreateAtomicSwap
	MsgRefundAtomicSwap = types.MsgRefundAtomicSwap
	Params              = types.Params
	QueryAssetSupply    = types.QueryAssetSupply
	QueryAtomicSwapByID = types.QueryAtomicSwapByID
	QueryAtomicSwaps    = types.QueryAtomicSwaps
	SwapDirection       = types.SwapDirection
	SwapStatus          = types.SwapStatus
)
