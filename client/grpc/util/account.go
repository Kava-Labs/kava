package util

import (
	"context"
	"fmt"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

// Account fetches an account via an address and returns the unpacked account
func (u *Util) Account(addr string) (authtypes.AccountI, error) {
	res, err := u.query.Auth.Account(context.Background(), &authtypes.QueryAccountRequest{
		Address: addr,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch account: %w", err)
	}

	var acc authtypes.AccountI
	err = u.encodingConfig.Marshaler.UnpackAny(res.Account, &acc)
	if err != nil {
		return nil, fmt.Errorf("failed to unpack account: %w", err)
	}
	return acc, nil
}

// BaseAccount fetches a base account via an address or returns an error if
// the account is not a base account
func (u *Util) BaseAccount(addr string) (authtypes.BaseAccount, error) {
	acc, err := u.Account(addr)
	if err != nil {
		return authtypes.BaseAccount{}, err
	}

	bAcc, ok := acc.(*authtypes.BaseAccount)
	if !ok {
		return authtypes.BaseAccount{}, fmt.Errorf("%s is not a base account", addr)
	}

	return *bAcc, nil
}
