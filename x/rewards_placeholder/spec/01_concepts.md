# Rewards implementation

## Part 1 - kavadist module

Module for minting kava incentive rewards

### Parameters

``` go
type Params struct {
  Active bool `json:"active" yaml:"active"`
  Periods  Periods `json:"periods" yaml:"periods"`
}

type Period struct {
  Start time.time `json:"start" yaml:"start"` // example "2020-03-01T15:20:00Z"
  End time.time `json:"end" yaml:"end"` // example "2020-06-01T15:20:00Z"
  Inflation sdk.Dec `json:"inflation" yaml:"inflation"` // example "0.10"  - 10% inflation
}

type Periods []Period
```

#### How Parameters are Used

Governance proposes an array of Periods, where each period stores a specified start and end date, and the inflation, expressed as a decimal representing the the yearly APR of KAVA tokens that will me minted during that period.

### Types

### Keys

* Stores the current period at certain prefix
* Stores the time since last mint at certain prefix

### Keeper

```go
type Keeper struct {
  sk supplyKeeper
  ...
}

// MintPeriodRewards mints the rewards for the period
func (k Keeper) MintPeriodRewards
  // calculates time since last period
  // calculates tokens to mint based on time since last period and the apr inflation rate
  // mints the tokens into the 'pool' module account

func (k Keeper) GetPeriod
func (k Keeper) SetPeriod
func (k Keeper) DeletePeriod
```

### BeginBlocker

In the begin blocker, `MintPeriodRewards` is called. If there is no period, the system is set to inactive. Also handles expired periods by deleting them and starting the next period, if there is one.

## Part 2 - kavarewards module

Module for creating and distributing time-locked KAVA rewards to users who create CDPs and mint USDX

### Parameters

```go
type Params struct {
  CollateralRewards
}

type CollateralReward struct {
  Denom string  // example bnb
  PeriodDuration int // example: 604800 (1 week)
  ClaimPeriodDuration int // how long eligible accounts can claim rewards // example: 1209600 two weeks
  Active bool // example: true
  Reward sdk.Coin // the number of coins distributed per period
  TimeLock int // example: 31540000 (1 year)
}
```

#### How Parameters are Used

Governance proposes an array of collateral rewards, with each item representing a collateral type that will be eligible for rewards. Governance can increase or decrease the number of coins awarded per period, the length of rewards periods, the length of claim periods, as well as adding or removing collaterals from rewards. All changes to parameters would take place in the _next_ period.

### Types

```go

// RewardPeriod stores the current reward period for the input denom
type RewardPeriod struct {
  Denom string
  Start time.Time
  End time.Time
  Reward sdk.Coin // per second reward payouts. For example, if we know from params that 10000KAVA is being paid out over 1 week (604800 rewards periods), then the value of reward would be (10000 * 1000000)/604800 = 16534ukava per second
  ClaimStart time.Time // only needed if there is a delay between rewards periods ending and claims periods starting
  ClaimEnd time.Time
  ClaimTimeLock int64 // the amount of time rewards are timelocked once they are sent to users
}

type ClaimPeriod struct {
  Denom string
  ID uint64
  End time.Time
}

// RewardClaim stores the claim on rewards that the input owner can redeem
type RewardClaim struct {
  Owner sdk.AccAddress
  Reward sdk.Coin
  ID unit64
}
```

### Messages

```go
// MsgClaimReward claims all active rewards that are owned by sender for the input denom
type MsgClaimReward struct {
  Sender sdk.AccAddress
  Denom string
}
```

### Keys

```go
// 0x01 <> []RewardPeriod array of the current active reward periods (max 1 reward period per denom)
// 0x02:Denom:ID <> ClaimPeriod object for that ID, indexed by denom and ID
// 0x03:Denom:ID:Owner <> RewardClaim object, indexed by Denom, ID and owner
// 0x04:denom <> NextClaimPeriodID the ID of the next claim period, indexed by denom
```

#### How Keys are Used

* Reward Period Creation - at genesis, or when a collateral is added to rewards, a `RewardPeriod` is created in the store by adding to the existing array of `[]RewardPeriod`. If the previous period for that collateral expired, it is deleted. This implies that, for each collateral, there will only ever be one reward period.

* Reward Period Deletion (and Claim Period Creation) - when a `RewardPeriod` expires, a new `ClaimPeriod` is created in the store with the next sequential ID for that collateral (ie, if the previous claim period was ID 1, the next one will be ID 2).

* Reward Claim Creation - Every block, CDPs are iterated over and the collateral denom is checked for rewards eligibility. For eligible cdps, a `RewardClaim` is created in the store for all CDP owners, if one doesn't already exist. The reward object is associated with a `ClaimPeriod` via the ID. This implies that `RewardClaim` are created before `ClaimPeriod` are created. Therefore, a user who submits a `MsgClaimReward` will only be paid out IF 1) they have one or more active `RewardClaim` objects, and 2) if the `ClaimPeriod` with the associated ID for that object exists AND the current block time  is between the start time and end time for that `ClaimPeriod`.

* Reward Claim Deletion (and Claim Period Deletion) - For claimed rewards, the `RewardClaim` is deleted from the store by deleting the key associated with that denom, ID, and owner. Unclaimed rewards are handled as follows: Each block, the `ClaimPeriod` objects for each denom are iterated over and checked for expiry. If expired, all `RewardClaim` objects for that ID are deleted, as well as the `ClaimPeriod` object. Since claim periods are monotonically increasing, once a non expired claim period is reached, the iteration can be stopped.

### Keeper

```go
type Keeper struct {
  supplyKeeper SupplyKeeper
  cdpKeeper CdpKeeper
  ...
}

// ApplyRewardsToCdpOwner creates a RewardClaim if one does not exist. Awards the ratable share of RewardPeriod.Rewards to the RewardClaim for the input owner.
func (k Keeper) ApplyRewardsToCdpOwner(ctx sdk.Context, recipient sdk.AccAddress, debt sdk.Coin, rewards sdk.Coin, collateralDenom string, principalDenom string, timeLock int64) {
  totalPrincipal := k.cdpKeeper.GetTotalPrincipal(ctx, collateralDenom).AmountOf(principalDenom)
  // calculate the fraction of debt minted by this cdp divided by the total amount of debt created for all cdps of that collateral type
  ratableDebtShare := debt.Amount.Quo(totalPrincipal) // note, implementation should use sdk.Dec
  // calculate the ratable portion of rewards that should be paid to recipient
  payoutAmount := ratableDebtShare.Mul(rewards.Amount) // note, implementation should use sdk.Dec and convert back to sdk.Int, then to sdk.Coins
  k.SendCoinsFromModuleToVestingAccount(ctx, types.ModuleName, recipient, payoutAmount, ctx.BlockTime(), ctx.BlockTime.Add(timeLock * time.Second))
}



// SendCoinsFromModuleToVestingAccount sends time-locked coins from the input module account to the recipient. If the recipients account is not a vesting account, it is converted to a periodic vesting account and the coins are added to the vesting balance as a vesting period with the input start and end times.
func (k Keeper) SendCoinsFromModuleToVestingAccount(ctx sdk.Context, senderModule string,
  recipientAddr sdk.AccAddress, amt sdk.Coins, startTime time.Time, endTime time.Time) {

  // Transfer to vesting account notes
  // Each vesting account looks like:
  PeriodicVestingAccount{
    StartTime: 1582818313 // int64
    EndTime: 1582918313 //int64
    Periods: Periods{Period{Length: 100, Amount: 100ukava}, Period{Length: 100, Amount : 100ukava}}
  }
  // where sum(length of all periods) = EndTime - StartTime

  // In the case that the vesting account DOES NOT exist, we can just created one with one period, with the length of the period being (endTime - startTime)
  // In the case that the vesting account DOES exist:
    // 1. If endTime is less than the vesting accounts end time, the vesting accounts periods can be ratably adjusted as follows:
      // 1. Any completed periods are unchanged
      // 2. Select the period whose end time is greater than or equal to the input end time. For example, if the input end time is 1582918313, find the first period with end time greater than or equal to 1582918313 by adding up the end times of the preceding periods.
      // 3a. If the selected period has end time GREATER than the input end time, insert a period such that the reward coins are paid out, then reduce the length of the old period by the length of the inserted period. For example, if we selected a period that had end time 1582928313 and length 200000, we would insert a period with end time 1582918313, and adjust the length of the selected period to be 10000. The original vesting coins MUST be updated.
      // 3b. If the selected period has end time EQUAL to the input end time, add the coins to that period. The original vesting coins MUST be updated.
      // 3c. If no period has an end time greater than or equal to the input end time, create a period with length equal to (inputEndtime - inputStartTime - (sum(length of all existing periods))). The end time of the vesting account MUST be updated. The original vesting coins MUST be updated.
  //  The same logic applies to validator vesting accounts, with the exception of case 3b. For that case, a new vesting period of length 0 (or 1 if 0 isn't valid) needs to be inserted with threshold 0, to signify that the vesting period has no validation requirement. Cases 3a and 3c still apply, with the additional requirement that the threshold 0 apply to the inserted/created period. Perhaps easier just to exclude validator vesting accounts and communicate that a fresh account should be used to any investors who are thinking about participating in rewards.
  }
```

### BeginBlocker

* Iterates over params to check if all active reward collaterals have active periods. A period is created if it does not exist.
* Iterates over all `RewardPeriod` to check if any are expired. If a period is expired, it is deleted, a new claim period is created, and a new reward period is created (or not) using the latest param info.
* For each non expired `RewardPeriod`, the time since the last block is used to determine the amount of rewards to pay out. Then, all cdps of that collateral type are iterated over and their `RewardClaim` objects are created/updated
* Iterates over all `ClaimPeriod` to check if any of them are expired. If a period is expired, all associated `RewardClaim` objects are deleted, as well as the `ClaimPeriod` object.
