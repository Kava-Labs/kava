package v0_11

import (
	"errors"
	"fmt"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Valid reward multipliers
const (
	Small  MultiplierName = "small"
	Medium MultiplierName = "medium"
	Large  MultiplierName = "large"
)

// GenesisClaimPeriodID stores the next claim id and its corresponding collateral type
type GenesisClaimPeriodID struct {
	CollateralType string `json:"collateral_type" yaml:"collateral_type"`
	ID             uint64 `json:"id" yaml:"id"`
}

// Validate performs a basic check of a GenesisClaimPeriodID fields.
func (gcp GenesisClaimPeriodID) Validate() error {
	if gcp.ID == 0 {
		return errors.New("genesis claim period id cannot be 0")
	}
	if strings.TrimSpace(gcp.CollateralType) == "" {
		return fmt.Errorf("collateral type cannot be blank: %v", gcp)
	}
	return nil
}

// GenesisClaimPeriodIDs array of GenesisClaimPeriodID
type GenesisClaimPeriodIDs []GenesisClaimPeriodID

// Validate checks if all the GenesisClaimPeriodIDs are valid and there are no duplicated
// entries.
func (gcps GenesisClaimPeriodIDs) Validate() error {
	seenIDS := make(map[string]bool)
	var key string
	for _, gcp := range gcps {
		key = gcp.CollateralType + string(gcp.ID)
		if seenIDS[key] {
			return fmt.Errorf("duplicated genesis claim period with id %d and collateral type %s", gcp.ID, gcp.CollateralType)
		}

		if err := gcp.Validate(); err != nil {
			return err
		}
		seenIDS[key] = true
	}

	return nil
}

// GenesisState is the state that must be provided at genesis.
type GenesisState struct {
	Params             Params                `json:"params" yaml:"params"`
	PreviousBlockTime  time.Time             `json:"previous_block_time" yaml:"previous_block_time"`
	RewardPeriods      RewardPeriods         `json:"reward_periods" yaml:"reward_periods"`
	ClaimPeriods       ClaimPeriods          `json:"claim_periods" yaml:"claim_periods"`
	Claims             Claims                `json:"claims" yaml:"claims"`
	NextClaimPeriodIDs GenesisClaimPeriodIDs `json:"next_claim_period_ids" yaml:"next_claim_period_ids"`
}

// NewGenesisState returns a new genesis state
func NewGenesisState(params Params, previousBlockTime time.Time, rp RewardPeriods, cp ClaimPeriods, c Claims, ids GenesisClaimPeriodIDs) GenesisState {
	return GenesisState{
		Params:             params,
		PreviousBlockTime:  previousBlockTime,
		RewardPeriods:      rp,
		ClaimPeriods:       cp,
		Claims:             c,
		NextClaimPeriodIDs: ids,
	}
}

// Validate performs basic validation of genesis data returning an
// error for any failed validation criteria.
func (gs GenesisState) Validate() error {
	if err := gs.Params.Validate(); err != nil {
		return err
	}
	if gs.PreviousBlockTime.IsZero() {
		return errors.New("previous block time cannot be 0")
	}
	if err := gs.RewardPeriods.Validate(); err != nil {
		return err
	}
	if err := gs.ClaimPeriods.Validate(); err != nil {
		return err
	}
	if err := gs.Claims.Validate(); err != nil {
		return err
	}
	return gs.NextClaimPeriodIDs.Validate()
}

// MultiplierName name for valid multiplier
type MultiplierName string

// IsValid checks if the input is one of the expected strings
func (mn MultiplierName) IsValid() error {
	switch mn {
	case Small, Medium, Large:
		return nil
	}
	return fmt.Errorf("invalid multiplier name: %s", mn)
}

// Params governance parameters for the incentive module
type Params struct {
	Active  bool    `json:"active" yaml:"active"` // top level governance switch to disable all rewards
	Rewards Rewards `json:"rewards" yaml:"rewards"`
}

// NewParams returns a new params object
func NewParams(active bool, rewards Rewards) Params {
	return Params{
		Active:  active,
		Rewards: rewards,
	}
}

// Validate checks that the parameters have valid values.
func (p Params) Validate() error {
	if err := validateActiveParam(p.Active); err != nil {
		return err
	}

	return validateRewardsParam(p.Rewards)
}

func validateActiveParam(i interface{}) error {
	_, ok := i.(bool)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	return nil
}

func validateRewardsParam(i interface{}) error {
	rewards, ok := i.(Rewards)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return rewards.Validate()
}

// Reward stores the specified state for a single reward period.
type Reward struct {
	Active           bool          `json:"active" yaml:"active"`                       // governance switch to disable a period
	CollateralType   string        `json:"collateral_type" yaml:"collateral_type"`     // the collateral type rewards apply to, must be found in the cdp collaterals
	AvailableRewards sdk.Coin      `json:"available_rewards" yaml:"available_rewards"` // the total amount of coins distributed per period
	Duration         time.Duration `json:"duration" yaml:"duration"`                   // the duration of the period
	ClaimMultipliers Multipliers   `json:"claim_multipliers" yaml:"claim_multipliers"` // the reward multiplier and timelock schedule - applied at the time users claim rewards
	ClaimDuration    time.Duration `json:"claim_duration" yaml:"claim_duration"`       // how long users have after the period ends to claim their rewards
}

// NewReward returns a new Reward
func NewReward(active bool, collateralType string, reward sdk.Coin, duration time.Duration, multiplier Multipliers, claimDuration time.Duration) Reward {
	return Reward{
		Active:           active,
		CollateralType:   collateralType,
		AvailableRewards: reward,
		Duration:         duration,
		ClaimMultipliers: multiplier,
		ClaimDuration:    claimDuration,
	}
}

// String implements fmt.Stringer
func (r Reward) String() string {
	return fmt.Sprintf(`Reward:
	Active: %t,
	CollateralType: %s,
	Available Rewards: %s,
	Duration: %s,
	%s,
	Claim Duration: %s`,
		r.Active, r.CollateralType, r.AvailableRewards, r.Duration, r.ClaimMultipliers, r.ClaimDuration)
}

// Validate performs a basic check of a reward fields.
func (r Reward) Validate() error {
	if !r.AvailableRewards.IsValid() {
		return fmt.Errorf("invalid reward coins %s for %s", r.AvailableRewards, r.CollateralType)
	}
	if !r.AvailableRewards.IsPositive() {
		return fmt.Errorf("reward amount must be positive, is %s for %s", r.AvailableRewards, r.CollateralType)
	}
	if r.Duration <= 0 {
		return fmt.Errorf("reward duration must be positive, is %s for %s", r.Duration, r.CollateralType)
	}
	if err := r.ClaimMultipliers.Validate(); err != nil {
		return err
	}
	if r.ClaimDuration <= 0 {
		return fmt.Errorf("claim duration must be positive, is %s for %s", r.ClaimDuration, r.CollateralType)
	}
	if strings.TrimSpace(r.CollateralType) == "" {
		return fmt.Errorf("collateral type cannot be blank: %s", r)
	}
	return nil
}

// Rewards array of Reward
type Rewards []Reward

// Validate checks if all the rewards are valid and there are no duplicated
// entries.
func (rs Rewards) Validate() error {
	rewardCollateralTypes := make(map[string]bool)
	for _, r := range rs {
		if rewardCollateralTypes[r.CollateralType] {
			return fmt.Errorf("cannot have duplicate reward collateral types: %s", r.CollateralType)
		}

		if err := r.Validate(); err != nil {
			return err
		}

		rewardCollateralTypes[r.CollateralType] = true
	}
	return nil
}

// String implements fmt.Stringer
func (rs Rewards) String() string {
	out := "Rewards\n"
	for _, r := range rs {
		out += fmt.Sprintf("%s\n", r)
	}
	return out
}

// Multiplier amount the claim rewards get increased by, along with how long the claim rewards are locked
type Multiplier struct {
	Name         MultiplierName `json:"name" yaml:"name"`
	MonthsLockup int64          `json:"months_lockup" yaml:"months_lockup"`
	Factor       sdk.Dec        `json:"factor" yaml:"factor"`
}

// NewMultiplier returns a new Multiplier
func NewMultiplier(name MultiplierName, lockup int64, factor sdk.Dec) Multiplier {
	return Multiplier{
		Name:         name,
		MonthsLockup: lockup,
		Factor:       factor,
	}
}

// Validate multiplier param
func (m Multiplier) Validate() error {
	if err := m.Name.IsValid(); err != nil {
		return err
	}
	if m.MonthsLockup < 0 {
		return fmt.Errorf("expected non-negative lockup, got %d", m.MonthsLockup)
	}
	if m.Factor.IsNegative() {
		return fmt.Errorf("expected non-negative factor, got %s", m.Factor.String())
	}

	return nil
}

// Multipliers slice of Multiplier
type Multipliers []Multiplier

// Validate validates each multiplier
func (ms Multipliers) Validate() error {
	for _, m := range ms {
		if err := m.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// String implements fmt.Stringer
func (ms Multipliers) String() string {
	out := "Claim Multipliers\n"
	for _, s := range ms {
		out += fmt.Sprintf("%s\n", s)
	}
	return out
}

// RewardPeriod stores the state of an ongoing reward
type RewardPeriod struct {
	CollateralType   string      `json:"collateral_type" yaml:"collateral_type"`
	Start            time.Time   `json:"start" yaml:"start"`
	End              time.Time   `json:"end" yaml:"end"`
	Reward           sdk.Coin    `json:"reward" yaml:"reward"` // per second reward payouts
	ClaimEnd         time.Time   `json:"claim_end" yaml:"claim_end"`
	ClaimMultipliers Multipliers `json:"claim_multipliers" yaml:"claim_multipliers"` // the reward multiplier and timelock schedule - applied at the time users claim rewards
}

// String implements fmt.Stringer
func (rp RewardPeriod) String() string {
	return fmt.Sprintf(`Reward Period:
	Collateral Type: %s,
	Start: %s,
	End: %s,
	Reward: %s,
	Claim End: %s,
	%s
	`, rp.CollateralType, rp.Start, rp.End, rp.Reward, rp.ClaimEnd, rp.ClaimMultipliers)
}

// NewRewardPeriod returns a new RewardPeriod
func NewRewardPeriod(collateralType string, start time.Time, end time.Time, reward sdk.Coin, claimEnd time.Time, claimMultipliers Multipliers) RewardPeriod {
	return RewardPeriod{
		CollateralType:   collateralType,
		Start:            start,
		End:              end,
		Reward:           reward,
		ClaimEnd:         claimEnd,
		ClaimMultipliers: claimMultipliers,
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
	if err := rp.ClaimMultipliers.Validate(); err != nil {
		return err
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
	CollateralType   string      `json:"collateral_type" yaml:"collateral_type"`
	ID               uint64      `json:"id" yaml:"id"`
	End              time.Time   `json:"end" yaml:"end"`
	ClaimMultipliers Multipliers `json:"claim_multipliers" yaml:"claim_multipliers"`
}

// NewClaimPeriod returns a new ClaimPeriod
func NewClaimPeriod(collateralType string, id uint64, end time.Time, multipliers Multipliers) ClaimPeriod {
	return ClaimPeriod{
		CollateralType:   collateralType,
		ID:               id,
		End:              end,
		ClaimMultipliers: multipliers,
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
	if err := cp.ClaimMultipliers.Validate(); err != nil {
		return err
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
	%s
	`, cp.CollateralType, cp.ID, cp.End, cp.ClaimMultipliers)
}

// GetMultiplier returns the named multiplier from the input claim period
func (cp ClaimPeriod) GetMultiplier(name MultiplierName) (Multiplier, bool) {
	for _, multiplier := range cp.ClaimMultipliers {
		if multiplier.Name == name {
			return multiplier, true
		}
	}
	return Multiplier{}, false
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
		CollateralType:   reward.CollateralType,
		Start:            blockTime,
		End:              blockTime.Add(reward.Duration),
		Reward:           rewardCoinPerSecond,
		ClaimEnd:         blockTime.Add(reward.Duration).Add(reward.ClaimDuration),
		ClaimMultipliers: reward.ClaimMultipliers,
	}
}
