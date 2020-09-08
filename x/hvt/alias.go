package hvt

import (
	"github.com/kava-labs/kava/x/hvt/keeper"
	"github.com/kava-labs/kava/x/hvt/types"
)

// nolint
// autogenerated code using github.com/rigelrozanski/multitool
// aliases generated for the following subdirectories:
// ALIASGEN: github.com/kava-labs/kava/x/hvt/types
// ALIASGEN: github.com/kava-labs/kava/x/hvt/keeper

const (
	EventTypeHarvest           = types.EventTypeHarvest
	EventTypeClaim             = types.EventTypeClaim
	EventTypeRewardPeriod      = types.EventTypeRewardPeriod
	EventTypeClaimPeriod       = types.EventTypeClaimPeriod
	EventTypeClaimPeriodExpiry = types.EventTypeClaimPeriodExpiry
	EventTypeLPDeposit         = types.EventTypeLPDeposit
	EventTypeGovDeposit        = types.EventTypeGovDeposit
	AttributeValueCategory     = types.AttributeValueCategory
	AttributeKeyDeposit        = types.AttributeKeyDeposit
	AttributeKeyClaimedBy      = types.AttributeKeyClaimedBy
	AttributeKeyClaimAmount    = types.AttributeKeyClaimAmount
	AttributeKeyRewardPeriod   = types.AttributeKeyRewardPeriod
	AttributeKeyClaimPeriod    = types.AttributeKeyClaimPeriod
	ModuleName                 = types.ModuleName
	LPAccount                  = types.LPAccount
	DelegatorAccount           = types.DelegatorAccount
	ModuleAccountName          = types.ModuleAccountName
	StoreKey                   = types.StoreKey
	RouterKey                  = types.RouterKey
	QuerierRoute               = types.QuerierRoute
	DefaultParamspace          = types.DefaultParamspace
	Small                      = types.Small
	Medium                     = types.Medium
	Large                      = types.Large
	LP                         = types.LP
	Stake                      = types.Stake
	QueryGetParams             = types.QueryGetParams
	QueryGetBalance            = types.QueryGetBalance
)

var (
	// functions aliases
	NewClaim                         = types.NewClaim
	RegisterCodec                    = types.RegisterCodec
	NewDeposit                       = types.NewDeposit
	NewGenesisState                  = types.NewGenesisState
	DefaultGenesisState              = types.DefaultGenesisState
	DepositKey                       = types.DepositKey
	DepositTypeIteratorKey           = types.DepositTypeIteratorKey
	ClaimKey                         = types.ClaimKey
	NewMsgDeposit                    = types.NewMsgDeposit
	NewMsgWithdraw                   = types.NewMsgWithdraw
	NewMsgClaimReward                = types.NewMsgClaimReward
	NewDistributionSchedule          = types.NewDistributionSchedule
	NewDelegatorDistributionSchedule = types.NewDelegatorDistributionSchedule
	NewMultiplier                    = types.NewMultiplier
	NewParams                        = types.NewParams
	DefaultParams                    = types.DefaultParams
	ParamKeyTable                    = types.ParamKeyTable
	NewPeriod                        = types.NewPeriod
	GetTotalVestingPeriodLength      = types.GetTotalVestingPeriodLength
	NewKeeper                        = keeper.NewKeeper
	NewQuerier                       = keeper.NewQuerier

	// variable aliases
	ModuleCdc                         = types.ModuleCdc
	ErrInvalidDepositDenom            = types.ErrInvalidDepositDenom
	ErrDepositNotFound                = types.ErrDepositNotFound
	ErrInvaliWithdrawAmount           = types.ErrInvaliWithdrawAmount
	ErrInvalidDepositType             = types.ErrInvalidDepositType
	ErrClaimNotFound                  = types.ErrClaimNotFound
	ErrZeroClaim                      = types.ErrZeroClaim
	ErrLPScheduleNotFound             = types.ErrLPScheduleNotFound
	ErrGovScheduleNotFound            = types.ErrGovScheduleNotFound
	ErrInvalidMultiplier              = types.ErrInvalidMultiplier
	ErrInsufficientModAccountBalance  = types.ErrInsufficientModAccountBalance
	ErrInvalidAccountType             = types.ErrInvalidAccountType
	ErrAccountNotFound                = types.ErrAccountNotFound
	PreviousBlockTimeKey              = types.PreviousBlockTimeKey
	PreviousDelegationDistributionKey = types.PreviousDelegationDistributionKey
	DepositsKeyPrefix                 = types.DepositsKeyPrefix
	ClaimsKeyPrefix                   = types.ClaimsKeyPrefix
	KeyActive                         = types.KeyActive
	KeyLPSchedules                    = types.KeyLPSchedules
	KeyDelegatorSchedule              = types.KeyDelegatorSchedule
	DefaultActive                     = types.DefaultActive
	DefaultGovSchedules               = types.DefaultGovSchedules
	DefaultLPSchedules                = types.DefaultLPSchedules
	DefaultDelegatorSchedules         = types.DefaultDelegatorSchedules
	DefaultPreviousBlockTime          = types.DefaultPreviousBlockTime
	GovDenom                          = types.GovDenom
)

type (
	Claim                          = types.Claim
	Deposit                        = types.Deposit
	GenesisState                   = types.GenesisState
	RewardMultiplier               = types.RewardMultiplier
	DepositType                    = types.DepositType
	MsgDeposit                     = types.MsgDeposit
	MsgWithdraw                    = types.MsgWithdraw
	MsgClaimReward                 = types.MsgClaimReward
	Params                         = types.Params
	DistributionSchedule           = types.DistributionSchedule
	DistributionSchedules          = types.DistributionSchedules
	DelegatorDistributionSchedule  = types.DelegatorDistributionSchedule
	DelegatorDistributionSchedules = types.DelegatorDistributionSchedules
	Multiplier                     = types.Multiplier
	Multipliers                    = types.Multipliers
	Keeper                         = keeper.Keeper
)
