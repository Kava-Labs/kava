package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	validatorvesting "github.com/kava-labs/kava/x/validator-vesting"
	"github.com/tendermint/tendermint/libs/cli"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authexported "github.com/cosmos/cosmos-sdk/x/auth/exported"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting"
	"github.com/cosmos/cosmos-sdk/x/genutil"
)

const (
	flagClientHome           = "home-client"
	flagVestingStart         = "vesting-start-time"
	flagVestingEnd           = "vesting-end-time"
	flagVestingAmt           = "vesting-amount"
	flagVestingPeriodsFile   = "vesting-periods-file"
	flagValidatorVestingFile = "validator-vesting-file"
)

// AddGenesisAccountCmd returns an add-genesis-account cobra Command.
func AddGenesisAccountCmd(
	ctx *server.Context, cdc *codec.Codec, defaultNodeHome, defaultClientHome string,
) *cobra.Command {

	cmd := &cobra.Command{
		Use:   "add-genesis-account [address_or_key_name] [coin][,[coin]]",
		Short: "Add a genesis account to genesis.json",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Add a genesis account to genesis.json. The provided account must specify
 the account address or key name and a list of initial coins. If a key name is given,
 the address will be looked up in the local Keybase. The list of initial tokens must
 contain valid denominations. Accounts may optionally be supplied with vesting parameters.
 If the account is a periodic or validator vesting account, vesting periods must be suppleid
 via a JSON file using the 'vesting-periods-file' flag or 'validator-vesting-file' flag,
 respectively.
 Example:
 %s add-genesis-account <account-name> <amount> --vesting-amount <amount> --vesting-end-time <unix-timestamp> --vesting-start-time <unix-timestamp> --vesting-periods <path/to/vesting.json>`, version.ClientName),
		),
		Args: cobra.ExactArgs(2),
		RunE: func(_ *cobra.Command, args []string) error {
			config := ctx.Config
			config.SetRoot(viper.GetString(cli.HomeFlag))

			addr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				// attempt to lookup address from Keybase if no address was provided
				kb, err := keys.NewKeyBaseFromDir(viper.GetString(flagClientHome))
				if err != nil {
					return err
				}

				info, err := kb.Get(args[0])
				if err != nil {
					return fmt.Errorf("failed to get address from Keybase: %w", err)
				}

				addr = info.GetAddress()
			}

			coins, err := sdk.ParseCoins(args[1])
			if err != nil {
				return fmt.Errorf("failed to parse coins: %w", err)
			}

			vestingStart := viper.GetInt64(flagVestingStart)
			vestingEnd := viper.GetInt64(flagVestingEnd)
			vestingAmt, err := sdk.ParseCoins(viper.GetString(flagVestingAmt))
			if err != nil {
				return fmt.Errorf("failed to parse vesting amount: %w", err)
			}
			vestingPeriodsFile := viper.GetString(flagVestingPeriodsFile)
			validatorVestingFile := viper.GetString(flagValidatorVestingFile)
			if vestingPeriodsFile != "" && validatorVestingFile != "" {
				return errors.New("Cannot specify both vesting-periods-file and validator-vesting-file")
			}

			// create concrete account type based on input parameters
			var genAccount authexported.GenesisAccount

			baseAccount := auth.NewBaseAccount(addr, coins.Sort(), nil, 0, 0)
			if !vestingAmt.IsZero() {
				baseVestingAccount, err := vesting.NewBaseVestingAccount(
					baseAccount, vestingAmt.Sort(), vestingEnd,
				)
				if err != nil {
					return fmt.Errorf("Failed to create base vesting account: %w", err)
				}

				switch {
				case vestingPeriodsFile != "":
					vestingPeriodsJSON, err := ParsePeriodicVestingJSON(cdc, vestingPeriodsFile)
					if err != nil {
						return fmt.Errorf("failed to parse periodic vesting account json file: %w", err)
					}
					genAccount = vesting.NewPeriodicVestingAccountRaw(baseVestingAccount, vestingStart, vestingPeriodsJSON.Periods)
				case validatorVestingFile != "":
					validatorVestingJSON, err := ParseValidatorVestingJSON(cdc, validatorVestingFile)
					if err != nil {
						return fmt.Errorf("failed to parse validator vesting account json file: %w", err)
					}
					consAddr, err := sdk.ConsAddressFromHex(validatorVestingJSON.ValidatorAddress)
					if err != nil {
						return fmt.Errorf("failed to convert validator address to bytes: %w", err)
					}
					genAccount = validatorvesting.NewValidatorVestingAccountRaw(baseVestingAccount, vestingStart, validatorVestingJSON.Periods, consAddr, validatorVestingJSON.ReturnAddress, validatorVestingJSON.SigningThreshold)
				case vestingStart != 0 && vestingEnd != 0:
					genAccount = vesting.NewContinuousVestingAccountRaw(baseVestingAccount, vestingStart)

				case vestingEnd != 0:
					genAccount = vesting.NewDelayedVestingAccountRaw(baseVestingAccount)

				default:
					return errors.New("invalid vesting parameters; must supply start and end time or end time")
				}
			} else {
				genAccount = baseAccount
			}

			if err := genAccount.Validate(); err != nil {
				return fmt.Errorf("failed to validate new genesis account: %w", err)
			}

			genFile := config.GenesisFile()
			appState, genDoc, err := genutil.GenesisStateFromGenFile(cdc, genFile)
			if err != nil {
				return fmt.Errorf("failed to unmarshal genesis state: %w", err)
			}

			authGenState := auth.GetGenesisStateFromAppState(cdc, appState)
			if authGenState.Accounts.Contains(addr) {
				return fmt.Errorf("cannot add account at existing address %s", addr)
			}

			// Add the new account to the set of genesis accounts and sanitize the
			// accounts afterwards.
			authGenState.Accounts = append(authGenState.Accounts, genAccount)
			authGenState.Accounts = auth.SanitizeGenesisAccounts(authGenState.Accounts)

			authGenStateBz, err := cdc.MarshalJSON(authGenState)
			if err != nil {
				return fmt.Errorf("failed to marshal auth genesis state: %w", err)
			}

			appState[auth.ModuleName] = authGenStateBz

			appStateJSON, err := cdc.MarshalJSON(appState)
			if err != nil {
				return fmt.Errorf("failed to marshal application genesis state: %w", err)
			}

			genDoc.AppState = appStateJSON
			return genutil.ExportGenesisFile(genDoc, genFile)
		},
	}

	cmd.Flags().String(cli.HomeFlag, defaultNodeHome, "node's home directory")
	cmd.Flags().String(flags.FlagKeyringBackend, flags.DefaultKeyringBackend, "Select keyring's backend (os|file|test)")
	cmd.Flags().String(flagClientHome, defaultClientHome, "client's home directory")
	cmd.Flags().String(flagVestingAmt, "", "amount of coins for vesting accounts")
	cmd.Flags().Uint64(flagVestingStart, 0, "schedule start time (unix epoch) for vesting accounts")
	cmd.Flags().Uint64(flagVestingEnd, 0, "schedule end time (unix epoch) for vesting accounts")
	cmd.Flags().String(flagVestingPeriodsFile, "", "path to file where periodic vesting schedule is specified")
	cmd.Flags().String(flagValidatorVestingFile, "", "path to file where validator vesting schedule is specified")
	return cmd
}

// ValidatorVestingJSON input json for validator-vesting-file flag
type ValidatorVestingJSON struct {
	Periods          vesting.Periods `json:"periods" yaml:"periods"`
	ValidatorAddress string          `json:"validator_address" yaml:"validator_address"`
	SigningThreshold int64           `json:"signing_threshold" yaml:"signing_threshold"`
	ReturnAddress    sdk.AccAddress  `json:"return_address,omitempty" yaml:"return_address,omitempty"`
}

// PeriodicVestingJSON input json for vesting-periods-file flag
type PeriodicVestingJSON struct {
	Periods vesting.Periods `json:"periods" yaml:"periods"`
}

// ParsePeriodicVestingJSON reads and parses ParsePeriodicVestingJSON from the file
func ParsePeriodicVestingJSON(cdc *codec.Codec, inputFile string) (PeriodicVestingJSON, error) {
	periodsInput := PeriodicVestingJSON{}

	content, err := ioutil.ReadFile(inputFile)

	if err != nil {
		return periodsInput, err
	}

	if err := cdc.UnmarshalJSON(content, &periodsInput); err != nil {
		return periodsInput, err
	}

	return periodsInput, nil
}

// ParseValidatorVestingJSON reads and parses ParseValidatorVestingJSON from the file
func ParseValidatorVestingJSON(cdc *codec.Codec, inputFile string) (ValidatorVestingJSON, error) {
	validatorVestingInput := ValidatorVestingJSON{}
	content, err := ioutil.ReadFile(inputFile)

	if err != nil {
		return validatorVestingInput, err
	}

	if err := cdc.UnmarshalJSON(content, &validatorVestingInput); err != nil {
		return validatorVestingInput, err
	}
	return validatorVestingInput, nil
}
