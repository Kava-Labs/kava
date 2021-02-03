package v0_9

import (
	"errors"
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	ModuleName = "incentive"
)

// GenesisClaimPeriodID stores the next claim id and its corresponding denom
type GenesisClaimPeriodID struct {
	Denom string `json:"denom" yaml:"denom"`
	ID    uint64 `json:"id" yaml:"id"`
}

// Validate performs a basic check of a GenesisClaimPeriodID fields.
func (gcp GenesisClaimPeriodID) Validate() error {
	if gcp.ID == 0 {
		return errors.New("genesis claim period id cannot be 0")
	}
	return sdk.ValidateDenom(gcp.Denom)
}

// GenesisClaimPeriodIDs array of GenesisClaimPeriodID
type GenesisClaimPeriodIDs []GenesisClaimPeriodID

// Validate checks if all the GenesisClaimPeriodIDs are valid and there are no duplicated
// entries.
func (gcps GenesisClaimPeriodIDs) Validate() error {
	seenIDS := make(map[string]bool)
	var key string
	for _, gcp := range gcps {
		key = gcp.Denom + fmt.Sprint(gcp.ID)
		if seenIDS[key] {
			return fmt.Errorf("duplicated genesis claim period with id %d and denom %s", gcp.ID, gcp.Denom)
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
	if gs.PreviousBlockTime.Unix() <= 0 {
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

// String implements fmt.Stringer
func (p Params) String() string {
	return fmt.Sprintf(`Params:
	Active: %t
	Rewards: %s`, p.Active, p.Rewards)
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
	Denom            string        `json:"denom" yaml:"denom"`                         // the collateral denom rewards apply to, must be found in the cdp collaterals
	AvailableRewards sdk.Coin      `json:"available_rewards" yaml:"available_rewards"` // the total amount of coins distributed per period
	Duration         time.Duration `json:"duration" yaml:"duration"`                   // the duration of the period
	TimeLock         time.Duration `json:"time_lock" yaml:"time_lock"`                 // how long rewards for this period are timelocked
	ClaimDuration    time.Duration `json:"claim_duration" yaml:"claim_duration"`       // how long users have after the period ends to claim their rewards
}

// NewReward returns a new Reward
func NewReward(active bool, denom string, reward sdk.Coin, duration time.Duration, timelock time.Duration, claimDuration time.Duration) Reward {
	return Reward{
		Active:           active,
		Denom:            denom,
		AvailableRewards: reward,
		Duration:         duration,
		TimeLock:         timelock,
		ClaimDuration:    claimDuration,
	}
}

// String implements fmt.Stringer
func (r Reward) String() string {
	return fmt.Sprintf(`Reward:
	Active: %t,
	Denom: %s,
	Available Rewards: %s,
	Duration: %s,
	Time Lock: %s,
	Claim Duration: %s`,
		r.Active, r.Denom, r.AvailableRewards, r.Duration, r.TimeLock, r.ClaimDuration)
}

// Validate performs a basic check of a reward fields.
func (r Reward) Validate() error {
	if !r.AvailableRewards.IsValid() {
		return fmt.Errorf("invalid reward coins %s for %s", r.AvailableRewards, r.Denom)
	}
	if !r.AvailableRewards.IsPositive() {
		return fmt.Errorf("reward amount must be positive, is %s for %s", r.AvailableRewards, r.Denom)
	}
	if r.Duration <= 0 {
		return fmt.Errorf("reward duration must be positive, is %s for %s", r.Duration, r.Denom)
	}
	if r.TimeLock < 0 {
		return fmt.Errorf("reward timelock must be non-negative, is %s for %s", r.TimeLock, r.Denom)
	}
	if r.ClaimDuration <= 0 {
		return fmt.Errorf("claim duration must be positive, is %s for %s", r.ClaimDuration, r.Denom)
	}
	return sdk.ValidateDenom(r.Denom)
}

// Rewards array of Reward
type Rewards []Reward

// Validate checks if all the rewards are valid and there are no duplicated
// entries.
func (rs Rewards) Validate() error {
	rewardDenoms := make(map[string]bool)
	for _, r := range rs {
		if rewardDenoms[r.Denom] {
			return fmt.Errorf("cannot have duplicate reward denoms: %s", r.Denom)
		}

		if err := r.Validate(); err != nil {
			return err
		}

		rewardDenoms[r.Denom] = true
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

// RewardPeriod stores the state of an ongoing reward
type RewardPeriod struct {
	Denom         string        `json:"denom" yaml:"denom"`
	Start         time.Time     `json:"start" yaml:"start"`
	End           time.Time     `json:"end" yaml:"end"`
	Reward        sdk.Coin      `json:"reward" yaml:"reward"` // per second reward payouts
	ClaimEnd      time.Time     `json:"claim_end" yaml:"claim_end"`
	ClaimTimeLock time.Duration `json:"claim_time_lock" yaml:"claim_time_lock"` // the amount of time rewards are timelocked once they are sent to users
}

// String implements fmt.Stringer
func (rp RewardPeriod) String() string {
	return fmt.Sprintf(`Reward Period:
	Denom: %s,
	Start: %s,
	End: %s,
	Reward: %s,
	Claim End: %s,
	Claim Time Lock: %s
	`, rp.Denom, rp.Start, rp.End, rp.Reward, rp.ClaimEnd, rp.ClaimTimeLock)
}

// NewRewardPeriod returns a new RewardPeriod
func NewRewardPeriod(denom string, start time.Time, end time.Time, reward sdk.Coin, claimEnd time.Time, claimTimeLock time.Duration) RewardPeriod {
	return RewardPeriod{
		Denom:         denom,
		Start:         start,
		End:           end,
		Reward:        reward,
		ClaimEnd:      claimEnd,
		ClaimTimeLock: claimTimeLock,
	}
}

// Validate performs a basic check of a RewardPeriod fields.
func (rp RewardPeriod) Validate() error {
	if rp.Start.Unix() <= 0 {
		return errors.New("reward period start time cannot be 0")
	}
	if rp.End.Unix() <= 0 {
		return errors.New("reward period end time cannot be 0")
	}
	if rp.Start.After(rp.End) {
		return fmt.Errorf("end period time %s cannot be before start time %s", rp.End, rp.Start)
	}
	if !rp.Reward.IsValid() {
		return fmt.Errorf("invalid reward amount: %s", rp.Reward)
	}
	if rp.ClaimEnd.Unix() <= 0 {
		return errors.New("reward period claim end time cannot be 0")
	}
	if rp.ClaimTimeLock == 0 {
		return errors.New("reward claim time lock cannot be 0")
	}
	return sdk.ValidateDenom(rp.Denom)
}

// RewardPeriods array of RewardPeriod
type RewardPeriods []RewardPeriod

// Validate checks if all the RewardPeriods are valid and there are no duplicated
// entries.
func (rps RewardPeriods) Validate() error {
	seenPeriods := make(map[string]bool)
	for _, rp := range rps {
		if seenPeriods[rp.Denom] {
			return fmt.Errorf("duplicated reward period with denom %s", rp.Denom)
		}

		if err := rp.Validate(); err != nil {
			return err
		}
		seenPeriods[rp.Denom] = true
	}

	return nil
}

// ClaimPeriod stores the state of an ongoing claim period
type ClaimPeriod struct {
	Denom    string        `json:"denom" yaml:"denom"`
	ID       uint64        `json:"id" yaml:"id"`
	End      time.Time     `json:"end" yaml:"end"`
	TimeLock time.Duration `json:"time_lock" yaml:"time_lock"`
}

// NewClaimPeriod returns a new ClaimPeriod
func NewClaimPeriod(denom string, id uint64, end time.Time, timeLock time.Duration) ClaimPeriod {
	return ClaimPeriod{
		Denom:    denom,
		ID:       id,
		End:      end,
		TimeLock: timeLock,
	}
}

// Validate performs a basic check of a ClaimPeriod fields.
func (cp ClaimPeriod) Validate() error {
	if cp.ID == 0 {
		return errors.New("claim period id cannot be 0")
	}
	if cp.End.Unix() <= 0 {
		return errors.New("claim period end time cannot be 0")
	}
	if cp.TimeLock == 0 {
		return errors.New("claim period time lock cannot be 0")
	}
	return sdk.ValidateDenom(cp.Denom)
}

// String implements fmt.Stringer
func (cp ClaimPeriod) String() string {
	return fmt.Sprintf(`Claim Period:
	Denom: %s,
	ID: %d,
	End: %s,
	Claim Time Lock: %s
	`, cp.Denom, cp.ID, cp.End, cp.TimeLock)
}

// ClaimPeriods array of ClaimPeriod
type ClaimPeriods []ClaimPeriod

// Validate checks if all the ClaimPeriods are valid and there are no duplicated
// entries.
func (cps ClaimPeriods) Validate() error {
	seenPeriods := make(map[string]bool)
	var key string
	for _, cp := range cps {
		key = cp.Denom + fmt.Sprint(cp.ID)
		if seenPeriods[key] {
			return fmt.Errorf("duplicated claim period with id %d and denom %s", cp.ID, cp.Denom)
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
	Owner         sdk.AccAddress `json:"owner" yaml:"owner"`
	Reward        sdk.Coin       `json:"reward" yaml:"reward"`
	Denom         string         `json:"denom" yaml:"denom"`
	ClaimPeriodID uint64         `json:"claim_period_id" yaml:"claim_period_id"`
}

// NewClaim returns a new Claim
func NewClaim(owner sdk.AccAddress, reward sdk.Coin, denom string, claimPeriodID uint64) Claim {
	return Claim{
		Owner:         owner,
		Reward:        reward,
		Denom:         denom,
		ClaimPeriodID: claimPeriodID,
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
	return sdk.ValidateDenom(c.Denom)
}

// String implements fmt.Stringer
func (c Claim) String() string {
	return fmt.Sprintf(`Claim:
	Owner: %s,
	Denom: %s,
	Reward: %s,
	Claim Period ID: %d,
	`, c.Owner, c.Denom, c.Reward, c.ClaimPeriodID)
}

// Claims array of Claim
type Claims []Claim

// Validate checks if all the claims are valid and there are no duplicated
// entries.
func (cs Claims) Validate() error {
	seemClaims := make(map[string]bool)
	var key string
	for _, c := range cs {
		key = c.Denom + fmt.Sprint(c.ClaimPeriodID) + c.Owner.String()
		if c.Owner != nil && seemClaims[key] {
			return fmt.Errorf("duplicated claim from owner %s and denom %s", c.Owner, c.Denom)
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
		Denom:         reward.Denom,
		Start:         blockTime,
		End:           blockTime.Add(reward.Duration),
		Reward:        rewardCoinPerSecond,
		ClaimEnd:      blockTime.Add(reward.Duration).Add(reward.ClaimDuration),
		ClaimTimeLock: reward.TimeLock,
	}
}
