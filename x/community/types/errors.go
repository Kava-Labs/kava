package types

import errorsmod "cosmossdk.io/errors"

var ErrInvalidParams = errorsmod.Register(ModuleName, 1, "invalid params")
