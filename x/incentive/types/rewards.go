package types

import (
	"errors"
	"fmt"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// RewardPeriod stores the state of an ongoing reward
type RewardPeriod struct {
	CollateralType string        `json:"collateral_type" yaml:"collateral_type"`
	Start          time.Time     `json:"start" yaml:"start"`
	End            time.Time     `json:"end" yaml:"end"`
	Reward         sdk.Coin      `json:"reward" yaml:"reward"` // per second reward payouts
	ClaimEnd       time.Time     `json:"claim_end" yaml:"claim_end"`
	ClaimTimeLock  time.Duration `json:"claim_time_lock" yaml:"claim_time_lock"` // the amount of time rewards are timelocked once they are sent to users
}

// String implements fmt.Stringer
func (rp RewardPeriod) String() string {
	return fmt.Sprintf(`Reward Period:
	Collateral Type: %s,
	Start: %s,
	End: %s,
	Reward: %s,
	Claim End: %s,
	Claim Time Lock: %s
	`, rp.CollateralType, rp.Start, rp.End, rp.Reward, rp.ClaimEnd, rp.ClaimTimeLock)
}

// NewRewardPeriod returns a new RewardPeriod
func NewRewardPeriod(collateralType string, start time.Time, end time.Time, reward sdk.Coin, claimEnd time.Time, claimTimeLock time.Duration) RewardPeriod {
	return RewardPeriod{
		CollateralType: collateralType,
		Start:          start,
		End:            end,
		Reward:         reward,
		ClaimEnd:       claimEnd,
		ClaimTimeLock:  claimTimeLock,
	}
}

// Validate performs a basic check of a RewardPeriod fields.
func (rp RewardPeriod) Validate() error {
	if rp.Start.IsZero() {
		return errors.New("reward period start time cannot be 0")
	}
	if rp.End.IsZero() {
		return errors.New("reward period end time cannot be 0")
	}
	if rp.Start.After(rp.End) {
		return fmt.Errorf("end period time %s cannot be before start time %s", rp.End, rp.Start)
	}
	if !rp.Reward.IsValid() {
		return fmt.Errorf("invalid reward amount: %s", rp.Reward)
	}
	if rp.ClaimEnd.IsZero() {
		return errors.New("reward period claim end time cannot be 0")
	}
	if rp.ClaimTimeLock == 0 {
		return errors.New("reward claim time lock cannot be 0")
	}
	if strings.TrimSpace(rp.CollateralType) == "" {
		return fmt.Errorf("reward period collateral type cannot be blank: %s", rp)
	}
	return nil
}

// RewardPeriods array of RewardPeriod
type RewardPeriods []RewardPeriod

// Validate checks if all the RewardPeriods are valid and there are no duplicated
// entries.
func (rps RewardPeriods) Validate() error {
	seenPeriods := make(map[string]bool)
	for _, rp := range rps {
		if seenPeriods[rp.CollateralType] {
			return fmt.Errorf("duplicated reward period with collateral type %s", rp.CollateralType)
		}

		if err := rp.Validate(); err != nil {
			return err
		}
		seenPeriods[rp.CollateralType] = true
	}

	return nil
}

// ClaimPeriod stores the state of an ongoing claim period
type ClaimPeriod struct {
	CollateralType string        `json:"collateral_type" yaml:"collateral_type"`
	ID             uint64        `json:"id" yaml:"id"`
	End            time.Time     `json:"end" yaml:"end"`
	TimeLock       time.Duration `json:"time_lock" yaml:"time_lock"`
}

// NewClaimPeriod returns a new ClaimPeriod
func NewClaimPeriod(collateralType string, id uint64, end time.Time, timeLock time.Duration) ClaimPeriod {
	return ClaimPeriod{
		CollateralType: collateralType,
		ID:             id,
		End:            end,
		TimeLock:       timeLock,
	}
}

// Validate performs a basic check of a ClaimPeriod fields.
func (cp ClaimPeriod) Validate() error {
	if cp.ID == 0 {
		return errors.New("claim period id cannot be 0")
	}
	if cp.End.IsZero() {
		return errors.New("claim period end time cannot be 0")
	}
	if cp.TimeLock == 0 {
		return errors.New("claim period time lock cannot be 0")
	}
	if strings.TrimSpace(cp.CollateralType) == "" {
		return fmt.Errorf("claim period collateral type cannot be blank: %s", cp)
	}
	return nil
}

// String implements fmt.Stringer
func (cp ClaimPeriod) String() string {
	return fmt.Sprintf(`Claim Period:
	Collateral Type: %s,
	ID: %d,
	End: %s,
	Claim Time Lock: %s
	`, cp.CollateralType, cp.ID, cp.End, cp.TimeLock)
}

// ClaimPeriods array of ClaimPeriod
type ClaimPeriods []ClaimPeriod

// Validate checks if all the ClaimPeriods are valid and there are no duplicated
// entries.
func (cps ClaimPeriods) Validate() error {
	seenPeriods := make(map[string]bool)
	var key string
	for _, cp := range cps {
		key = cp.CollateralType + string(cp.ID)
		if seenPeriods[key] {
			return fmt.Errorf("duplicated claim period with id %d and collateral type %s", cp.ID, cp.CollateralType)
		}

		if err := cp.Validate(); err != nil {
			return err
		}
		seenPeriods[key] = true
	}

	return nil
}

// Claim stores the rewards that can be claimed by owner
type Claim struct {
	Owner          sdk.AccAddress `json:"owner" yaml:"owner"`
	Reward         sdk.Coin       `json:"reward" yaml:"reward"`
	CollateralType string         `json:"collateral_type" yaml:"collateral_type"`
	ClaimPeriodID  uint64         `json:"claim_period_id" yaml:"claim_period_id"`
}

// NewClaim returns a new Claim
func NewClaim(owner sdk.AccAddress, reward sdk.Coin, collateralType string, claimPeriodID uint64) Claim {
	return Claim{
		Owner:          owner,
		Reward:         reward,
		CollateralType: collateralType,
		ClaimPeriodID:  claimPeriodID,
	}
}

// Validate performs a basic check of a Claim fields.
func (c Claim) Validate() error {
	if c.Owner.Empty() {
		return errors.New("claim owner cannot be empty")
	}
	if !c.Reward.IsValid() {
		return fmt.Errorf("invalid reward amount: %s", c.Reward)
	}
	if c.ClaimPeriodID == 0 {
		return errors.New("claim period id cannot be 0")
	}
	if strings.TrimSpace(c.CollateralType) == "" {
		return fmt.Errorf("claim collateral type cannot be blank: %s", c)
	}
	return nil
}

// String implements fmt.Stringer
func (c Claim) String() string {
	return fmt.Sprintf(`Claim:
	Owner: %s,
	Collateral Type: %s,
	Reward: %s,
	Claim Period ID: %d,
	`, c.Owner, c.CollateralType, c.Reward, c.ClaimPeriodID)
}

// Claims array of Claim
type Claims []Claim

// AugmentedClaim claim type with information about whether the claim is currently claimable
type AugmentedClaim struct {
	Claim Claim

	Claimable bool
}

func (ac AugmentedClaim) String() string {
	return fmt.Sprintf(`Claim:
	Owner: %s,
	Collateral Type: %s,
	Reward: %s,
	Claim Period ID: %d,
	Claimable: %t,
	`, ac.Claim.Owner, ac.Claim.CollateralType, ac.Claim.Reward, ac.Claim.ClaimPeriodID, ac.Claimable,
	)
}

// NewAugmentedClaim returns a new augmented claim struct
func NewAugmentedClaim(claim Claim, claimable bool) AugmentedClaim {
	return AugmentedClaim{
		Claim:     claim,
		Claimable: claimable,
	}
}

// AugmentedClaims array of AugmentedClaim
type AugmentedClaims []AugmentedClaim

// Validate checks if all the claims are valid and there are no duplicated
// entries.
func (cs Claims) Validate() error {
	seemClaims := make(map[string]bool)
	var key string
	for _, c := range cs {
		key = c.CollateralType + string(c.ClaimPeriodID) + c.Owner.String()
		if c.Owner != nil && seemClaims[key] {
			return fmt.Errorf("duplicated claim from owner %s and collateral type %s", c.Owner, c.CollateralType)
		}

		if err := c.Validate(); err != nil {
			return err
		}
		seemClaims[key] = true
	}

	return nil
}

// NewRewardPeriodFromReward returns a new reward period from the input reward and block time
func NewRewardPeriodFromReward(reward Reward, blockTime time.Time) RewardPeriod {
	// note: reward periods store the amount of rewards paid PER SECOND
	rewardsPerSecond := sdk.NewDecFromInt(reward.AvailableRewards.Amount).Quo(sdk.NewDecFromInt(sdk.NewInt(int64(reward.Duration.Seconds())))).TruncateInt()
	rewardCoinPerSecond := sdk.NewCoin(reward.AvailableRewards.Denom, rewardsPerSecond)
	return RewardPeriod{
		CollateralType: reward.CollateralType,
		Start:          blockTime,
		End:            blockTime.Add(reward.Duration),
		Reward:         rewardCoinPerSecond,
		ClaimEnd:       blockTime.Add(reward.Duration).Add(reward.ClaimDuration),
		ClaimTimeLock:  reward.TimeLock,
	}
}
