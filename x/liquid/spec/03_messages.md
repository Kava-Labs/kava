<!--
order: 3
-->

# Messages

`bkava` is minted using `MsgMintDerivative`.


```go
// MsgMintDerivative defines the Msg/MintDerivative request type.
type MsgMintDerivative struct {
	// sender is the owner of the delegation to be converted
	Sender string `protobuf:"bytes,1,opt,name=sender,proto3" json:"sender,omitempty"`
	// validator is the validator of the delegation to be converted
	Validator string `protobuf:"bytes,2,opt,name=validator,proto3" json:"validator,omitempty"`
	// amount is the quantity of staked assets to be converted
	Amount types.Coin `protobuf:"bytes,3,opt,name=amount,proto3" json:"amount"`
}
```

### Actions

* converts an existing delegation into bkava tokens
* delegation is transferred from the sender to a module account
* validator specific bkava are minted and sent to the sender

### Example:

```jsonc
{
  // user who owns the delegation
  "sender": "kava10wlnqzyss4accfqmyxwx5jy5x9nfkwh6qm7n4t",
  // validator the user has delegated to
  "validator": "kavavaloper1ypjp0m04pyp73hwgtc0dgkx0e9rrydeckewa42",
  // amount of staked ukava to be converted into bkava
  "amount": {
    "amount": "1000000000",
    "denom": "ukava"
  }
}
```

`bkava` can be burned using `MsgBurnDerivative`.

```go
// MsgBurnDerivative defines the Msg/BurnDerivative request type.
type MsgBurnDerivative struct {
	// sender is the owner of the derivatives to be converted
	Sender string `protobuf:"bytes,1,opt,name=sender,proto3" json:"sender,omitempty"`
	// validator is the validator of the derivatives to be converted
	Validator string `protobuf:"bytes,2,opt,name=validator,proto3" json:"validator,omitempty"`
	// amount is the quantity of derivatives to be converted
	Amount types.Coin `protobuf:"bytes,3,opt,name=amount,proto3" json:"amount"`
}
```

### Actions

* converts bkava tokens into a delegation
* bkava is burned
* a delegation equal to number of bkava is transferred to user


### Example

```jsonc
{
  // user who owns the bkava
  "sender": "kava10wlnqzyss4accfqmyxwx5jy5x9nfkwh6qm7n4t",
  // the amount of bkava the user wants to convert back into normal staked kava
  "amount": {
    "amount": "1234000000",
    "denom": "bkava-kavavaloper1ypjp0m04pyp73hwgtc0dgkx0e9rrydeckewa42"
  },
  // the validator behind the bkava, this address must match the one embedded in the bkava denom above
  "validator": "kavavaloper1ypjp0m04pyp73hwgtc0dgkx0e9rrydeckewa42"
}
```
