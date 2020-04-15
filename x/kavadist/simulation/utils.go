package simulation

import (
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TODO use the official version available in v0.38.2 of the sdk
// https://github.com/cosmos/cosmos-sdk/blob/e8d89a2fe26175b73545a3e79ae783032b4e975e/types/decimal.go#L328
func approxRoot(d sdk.Dec, root uint64) (guess sdk.Dec, err error) {
	defer func() {
		if r := recover(); r != nil {
			var ok bool
			err, ok = r.(error)
			if !ok {
				err = errors.New("out of bounds")
			}
		}
	}()

	if d.IsNegative() {
		absRoot, err := approxRoot(d.MulInt64(-1), root)
		return absRoot.MulInt64(-1), err
	}
	if root == 1 || d.IsZero() || d.Equal(sdk.OneDec()) {
		return d, nil
	}
	if root == 0 {
		return sdk.OneDec(), nil
	}
	rootInt := sdk.NewInt(int64(root))
	guess, delta := sdk.OneDec(), sdk.OneDec()
	for delta.Abs().GT(sdk.SmallestDec()) {
		prev := power(guess, (root - 1))
		if prev.IsZero() {
			prev = sdk.SmallestDec()
		}
		delta = d.Quo(prev)
		delta = delta.Sub(guess)
		delta = delta.QuoInt(rootInt)

		guess = guess.Add(delta)
	}
	return guess, nil
}

// Power returns a the result of raising to a positive integer power
func power(d sdk.Dec, power uint64) sdk.Dec {
	if power == 0 {
		return sdk.OneDec()
	}
	tmp := sdk.OneDec()
	for i := power; i > 1; {
		if i%2 == 0 {
			i /= 2
		} else {
			tmp = tmp.Mul(d)
			i = (i - 1) / 2
		}
		d = d.Mul(d)
	}
	return d.Mul(tmp)
}
