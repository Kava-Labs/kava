# `x/precisebank`

## Abstract

This document specifies the precisebank module of Kava.

The precisebank module is responsible for extending the precision of `x/bank`, intended to be used for the `x/evm`. It serves as a wrapper of `x/bank` to increase the precision of KAVA from 6 to 18 decimals, while preserving the behavior of existing `x/bank` balances.

This module is used only by `x/evm` where 18 decimal points are expected.

## Contents

- [Background](#background)
  - [Adding](#adding)
  - [Subtracting](#subtracting)
  - [Transfer](#transfer)
    - [Setup](#setup)
    - [Remainder does not change](#remainder-does-not-change)
    - [Reserve](#reserve)
  - [Burn](#burn)
  - [Mint](#mint)
- [State](#state)
- [Keepers](#keepers)
- [Messages](#messages)
- [Events](#events)
  - [Keeper Events](#keeper-events)
    - [SendCoins](#sendcoins)
    - [MintCoins](#mintcoins)
    - [BurnCoins](#burncoins)
- [Client](#client)
  - [gRPC](#grpc)
    - [TotalFractionalBalances](#totalfractionalbalances)
    - [Remainder](#remainder)
    - [FractionalBalance](#fractionalbalance)

## Background

The standard unit of currency on the Kava Chain is `KAVA`.  This is denominated by the atomic unit `ukava`, which represents $10^{-6}$ `KAVA` and there are $10^6$ `ukava` per `KAVA`.

In order to support 18 decimals of precision while maintaining `ukava` as the cosmos-native atomic unit, we further split each `ukava` unit into $10^{12}$ `akava` units, the native currency of the Kava EVM.

This gives a full $10^{18}$ precision on the EVM. In order to avoid confusion with atomic `ukava` units, we will refer to `akava` as "sub-atomic units".

To review we have:
 - `ukava`, the cosmos-native unit and atomic unit of the Kava chain
 - `akava`, the evm-native unit and sub-atomic unit of the Kava chain

In order to maintain consistency between the `akava` supply and the `ukava` supply, we add the constraint that each sub-atomic `akava`, may only exist as part of an atomic `ukava`. Every `akava` is fully backed by a `ukava` in the `x/bank` module.

This is a requirement since `ukava` balances in `x/bank` are shared between the cosmos modules and the EVM.  We are wrapping and extending the `x/bank` module with the `x/precisebank` module to add an extra $10^{12}$ units of precision.  If $10^{12}$ `akava` is transferred in the EVM, the cosmos modules will see a 1 `ukava` transfer and vice versa.  If `akava` was not fully backed by `ukava`, then balance changes would not be fully consistent across the cosmos and the EVM.

This brings us to how account balances are extended to represent `akava` balances larger than $10^{12}$.  First, we define $a(n)$, $b(n)$, and $C$ where $a(n)$ is the `akava` balance of account `n`, $b(n)$ is the `ukava` balance of account `n` stored in the `x/bank` module, and $C$ is the conversion factor equal to $10^{12}$.

Any $a(n)$ divisible by $C$, can be represented by $C$ * $b(n)$.  Any remainder not divisible by $C$, we define the "fractional balance" as $f(n)$ and store this in the `x/precisebank` store.

Thus,

$$a(n) = b(n) \cdot C + f(n)$$

where

$$0 \le f(n) < C$$

$$a(n), b(n) \ge 0$$

This is the quotient-remainder theorem and any $a(n)$ can be represented by unique integers $b(n)$, $f(n)$ where

$$b(n) = \lfloor a(n)/C \rfloor$$

$$f(n) = a(n)\bmod{C}$$

With this definition in mind we will refer to $b(n)$ units as integer units, and $f(n)$ as fractional units.

Now since $f(n)$ is stored in the `x/precisebank` and not tracked by the `x/bank` keeper, these are not counted in the `ukava` supply, so if we define

$$T_a \equiv \sum_{n \in \mathcal{A}}{a(n)}$$

$$T_b \equiv \sum_{n \in \mathcal{A}}{b(n)}$$

where $\mathcal{A}$ is the set of all accounts, $T_a$ is the total `akava` supply, and $T_b$ is the total `ukava` supply, then a reserve account $R$ is added such that

$$a(R) = 0$$

$$b(R) \cdot C = \sum_{n \in \mathcal{A}}{f(n)} + r$$

where $R$ is the module account of the `x/precisebank`, and $r$ is the remainder or fractional amount backed by $b(R)$, but not yet in circulation such that

$$T_a = T_b \cdot C - r$$

and

$$ 0 <= r < C$$

We see that $0 \le T_b \cdot C - T_a < C$. If we mint, burn, or transfer `akava` such that this inequality would be invalid after updates to account balances, we adjust the $T_b$ supply by minting or burning to the reserve account which holds `ukava` equal to that of all `akava` balances less than `C` plus the remainder.

If we didn't add these constraints, then the total supply of `ukava` reported by the bank keeper would not account for the `akava` units.  We would incorrectly increase the supply of `akava` without increasing the reported total supply of KAVA.

### Adding

When adding we have

$$a'(n) = a(n) + a$$

$$b'(n) \cdot C + f'(n) = b(n) \cdot C + f(n) + a$$

where $a'(n)$ is the new `akava` balance after adding `akava` amount $a$. These
must hold true for all $a$. We can determine the new $b'(n)$ and $f'(n)$ with the following formula.

$$f'(n) = f(n) + a \mod{C}$$

$$b'(n) = \begin{cases} b(n) + \lfloor a/C \rfloor & f'(n) \geq f(n) \\
b(n) + \lfloor a/C \rfloor + 1 & f'(n) < f(n) \end{cases}$$

We can see that $b'(n)$ is incremented by an additional 1 integer unit if
$f'(n) < f(n)$ because the new balance requires an arithmetic carry from the
fractional to the integer unit.

### Subtracting

When subtracting we have

$$a'(n) = a(n) - a$$

$$b'(n) \cdot C + f'(n) = b(n) \cdot C + f(n) - a$$

and

$$f'(n) = f(n) - a \mod{C}$$

$$b'(n) = \begin{cases} b(n) - \lfloor a/C \rfloor & f'(n) \leq f(n) \\
b(n) - \lfloor a/C \rfloor - 1 & f'(n) > f(n) \end{cases}$$

Similar to the adding case, we subtract $b'(n)$ by an additional 1 if
$f'(n) > f(n)$ because $f(n)$ is insufficient on its own and requires an
arithmetic borrow from the integer units.

### Transfer

A transfer is a combination of adding and subtracting of a single amount between
two different accounts. The transfer is valid if both the subtraction for the
sender and the addition for the receiver are valid.

#### Setup

Let two accounts $1$ and $2$ have balances $a(1)$ and $a(2)$, and $a$ is the
amount to transfer. Assume that $a(1) \ge a$ to ensure that the transfer is
valid. We initiate a transfer by subtracting $a$ from account $1$ and adding $a$
to account $2$, yielding

$$a'(1) = a(1) - a$$

$$a'(2) = a(2) + a$$

The reserve account must also be updated to reflect the change in the total
supply of fractional units.

$$b(R) \cdot C = \sum_{n \in \mathcal{A}}{f(n)} + r$$

$$b'(R) \cdot C = \sum_{n \in \mathcal{A}}{f'(n)} + r'$$

With these two formulas, we can determine the new remainder and reserve by using
the delta of the sum of fractional units and the remainder.

$$(b'(R)-b(R)) \cdot C = \sum_{n \in \mathcal{A}}{f'(n)} - \sum_{n \in \mathcal{A}}{f(n)} + r' - r$$

Since only two accounts are involved in the transfer, we can use the two account
balances in place of the fractional sum delta.

$$(b'(R)-b(R)) \cdot C = f'(1) - f(1) + f'(2) - f(2) + r' - r$$

#### Remainder does not change

Take $\mod{C}$ of both sides of the equation.

$$(b'(R)-b(R)) \cdot C \mod{C} = [f'(1) - f(1) + f'(2) - f(2) + r' - r] \mod{C}$$

Since $C$ is a multiple of $C$, the left side of the equation is $0$.

$$0 = f'(1) - f(1) + f'(2) - f(2) + r' - r \mod{C}$$

Replace $f'(1)$ and $f'(2)$ with their definitions in terms of $f(1)$ and $f(2)$.

$$0 = (f(1) - a)\bmod{C} - f(1) + (f(2) + a)\bmod{C} - f(2) + r' - r \mod{C}$$

This can be simplified to:

$$0 = f(1) - a - f(1) + f(2) + a - f(2) + r' - r \mod{C}$$

Canceling out terms $a$, $f(1)$ and $f(2)$.

$$0 = r' - r \mod{C}$$

By the quotient remainder theorem, we can express $r' - r$ as:

$$q * C = r' - r$$

for some integer $q$.

With our known range of $r$ and $r'$:

$$0 \leq r' < C, 0 \leq r < C$$

We can see that $r' - r$ must be in the range

$$ -C < r' - r < C$$

This implies that $q$ must be $0$ as there is no other integer $q$ that satisfies the inequality.

$$ -C < q * C < C$$

$$q = 0$$

$$ r' - r = 0$$

Therefore, the remainder does not change during a transfer.

#### Reserve

The reserve account must be updated to reflect the change in the total supply of fractional units.

The change in reserve is determined by the change in the fractional units of the two accounts.

$$(b'(R)-b(R)) \cdot C = f'(1) - f(1) + f'(2) - f(2)$$

For $f'(1)$, we can represent the new fractional balance as:

$$f'(1) = f(1) - a \mod{C}$$

$$f'(1)\bmod{C}= f(1)\bmod{C} - a \bmod{C} \mod{C}$$

$$f'(1) = f(1) - a \bmod{C} \mod{C}$$

This results in two cases for $f'(1)$:

$$f'(1) = \begin{cases} f(1) - a\bmod{C} & 0 \leq f(1) - a\bmod{C} \\
f(1) - a\bmod{C} + C & 0 > f(1) - a\bmod{C} \end{cases}$$

Since we can identify the following:

$$f'(1) \leq f(1) \Longleftrightarrow  f'(1) = f(1) - a\bmod{C} $$

$$f'(1) > f(1) \Longleftrightarrow  f'(1) = f(1) - a\bmod{C} + C$$

We can simplify the two cases for $f'(1)$:

$$f'(1) = \begin{cases} f(1) - a\bmod{C} & f'(1) \leq f(1) \\
f(1) - a\bmod{C} + C & f'(1) > f(1) \end{cases}$$

The same for $f'(2)$:

$$f'(2) = f(2) + a \mod{C}$$

$$f'(2)\bmod{C}= f(2)\bmod{C} + a \bmod{C} \mod{C}$$

$$f'(2) = f(2) + a \bmod{C} \mod{C}$$

$$f'(2) = \begin{cases} f(2) + a\bmod{C} & f'(2) \geq f(2) \\
f(2) + a\bmod{C} - C & f'(2) < f(2) \end{cases}$$

Bringing the two cases for the two accounts together to determine the change in the reserve account:

$$b'(R) - b(R) \cdot C = \begin{cases} f(1) - a\bmod{C} + C - f(1) + f(2) + a\bmod{C} - C + f(2) & f'(1) > f(1) \land f'(2) < f(2) \\
f(1) - a\bmod{C} - f(1) + f(2) + a\bmod{C} - C + f(2) & f'(1) \leq f(1) \land f'(2) < f(2) \\
f(1) - a\bmod{C} + C - f(1) + f(2) + a\bmod{C} + f(2) & f'(1) > f(1) \land f'(2) \geq f(2) \\
f(1) - a\bmod{C} - f(1) + f(2) + a\bmod{C} + f(2) & f'(1) \leq f(1) \land f'(2) \geq f(2) \\
\end{cases}$$

This simplifies to:

$$b'(R) - b(R) \cdot C = \begin{cases} 0 & f'(1) > f(1) \land f'(2) < f(2) \\
-C & f'(1) \leq f(1) \land f'(2) < f(2) \\
C & f'(1) > f(1) \land f'(2) \geq f(2) \\
0 & f'(1) \leq f(1) \land f'(2) \geq f(2) \\
\end{cases}$$

Simplifying further by dividing by $C$:

$$b'(R) - b(R) = \begin{cases} 0 & f'(1) > f(1) \land f'(2) < f(2) \\
-1 & f'(1) \leq f(1) \land f'(2) < f(2) \\
1 & f'(1) > f(1) \land f'(2) \geq f(2) \\
0 & f'(1) \leq f(1) \land f'(2) \geq f(2) \\
\end{cases}$$

Thus the reserve account is updated based on the changes in the fractional units of the two accounts.

### Burn

When burning, we change only 1 account. Assume we are burning an amount $a$ from account $1$.

$$a'(1) = a(1) - a$$

The change in reserve is determined by the change in the fractional units of the account and the remainder.

$$(b'(R)-b(R)) \cdot C = f'(1) - f(1) + r' - r$$

The new fractional balance is:

$$f'(1) = f(1) - a \mod{C}$$

Apply modulo $C$ to both sides of the equation.

$$f'(1)\bmod{C}= f(1)\bmod{C} - a \bmod{C} \mod{C}$$

This simplifies to:

$$f'(1) = f(1) - a \bmod{C} \mod{C}$$

We can see two cases for $f'(1)$, depending on whether the new fractional balance is less than the old fractional balance.

$$f'(1) = \begin{cases} f(1) - a\bmod{C} & f'(1) \leq f(1) \\
f(1) - a\bmod{C} + C & f'(1) > f(1) \end{cases}$$

The second case occurs when we need to borrow from the integer units.

We update the remainder by adding $a$ to $r$ as burning increases the amount no longer in circulation but still backed by the reserve.

$$r' = r + a \mod{C}$$

$$r'\bmod{C}= r\bmod{C} + a \bmod{C} \mod{C}$$

$$r' = r + a \bmod{C} \mod{C}$$

We can see two cases for $r'$, depending on whether the new remainder is less than the old remainder.

$$r' = \begin{cases} r + a\bmod{C} & r' \geq r \\
r + a\bmod{C} - C & r' < r \end{cases}$$

The reserve account is updated based on the changes in the fractional units of the account and remainder.

$$b'(R) - b(R) = \begin{cases} 0 & f'(1) > f(1) \land r' < r \\
-1 & f'(1) \leq f(1) \land r' < r \\
1 & f'(1) > f(1) \land r' \geq r \\
0 & f'(1) \leq f(1) \land r' \geq r \\
\end{cases}$$

### Mint

Minting is similar to burning, but we add to the account instead of
removing it. Assume we are minting an amount $a$ to account $1$.

$$a'(1) = a(1) + a$$

The change in reserve is determined by the change in the fractional units of the account and the remainder.

$$(b'(R)-b(R)) \cdot C = f'(1) - f(1) + r' - r$$

The new fractional balance is:

$$f'(1) = f(1) + a \mod{C}$$

Apply modulo $C$ to both sides of the equation.

$$f'(1)\bmod{C}= f(1)\bmod{C} + a \bmod{C} \mod{C}$$

$$f'(1) = f(1) + a \bmod{C} \mod{C}$$

We can see two cases for $f'(1)$, depending on whether the new fractional balance is greater than the old fractional balance.

$$f'(1) = \begin{cases} f(1) + a\bmod{C} & f'(1) \geq f(1) \\
f(1) + a\bmod{C} - C & f'(1) < f(1) \end{cases}$$

The second case occurs when we need to carry to the integer unit.

We update the remainder by subtracting $a$ from $r$ as minting decreases the amount no longer in circulation but still backed by the reserve.

$$r' = r - a \mod{C}$$

$$r'\bmod{C}= r\bmod{C} - a \bmod{C} \mod{C}$$

$$r' = r - a \bmod{C} \mod{C}$$

$$r' = \begin{cases} r - a\bmod{C} & r' \leq r \\
r - a\bmod{C} + C & r' > r \end{cases}$$

The reserve account is updated based on the changes in the fractional units of the account and the remainder.

$$b'(R) - b(R) = \begin{cases} 0 & r' > r \land f'(1) < f(1) \\
-1 & r' \leq r \land f'(1) < f(1) \\
1 & r' > r \land f'(1) \geq f(1) \\
0 & r' \leq r \land f'(1) \geq f(1) \\
\end{cases}$$

## State

The `x/precisebank` module keeps state of the following:
1. Account fractional balances.
2. Remainder amount. This amount represents the fractional amount that is backed
   by the reserve account but not yet in circulation. This can be non-zero if
   a fractional amount less than `1ukava` is minted.

   **Note:** Currently, mint and burns are only used to transfer fractional
   amounts between accounts via `x/evm`. This means mint and burns on mainnet
   state will always be equal and opposite, always resulting in a zero remainder
   at the end of each transaction and block.

The `x/precisebank` module does not keep track of the reserve as it is stored in
the `x/bank` module.

## Keepers

The `x/precisebank module only exposes one keeper that wraps the bank module`
keeper and implements bank keeper compatible methods to support extended coin.
This complies with the `x/evm` module interface for `BankKeeper`.

```go
type BankKeeper interface {
	authtypes.BankKeeper
	SpendableCoin(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin
	SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	MintCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error
	BurnCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error
}
```

## Messages

The `x/precisebank` module does not have any messages and is intended to be used
by other modules as a replacement of the bank module.

## Events

### Keeper Events

The `x/precisebank` module emits the following events, that are meant to be
match the events emitted by the `x/bank` module. Events emitted by
`x/precisebank` will only contain `akava` amounts, as the `x/bank` module will
emit events with all other denoms. This means if an account transfers multiple
coins including `akava`, the `x/precisebank` module will emit an event with the
full `akava` amount. If `ukava` is included in a transfer, mint, or burn, the
`x/precisebank` module will emit an event with the full equivalent `akava`
amount.

#### SendCoins

```json
{
  "type": "transfer",
  "attributes": [
    {
      "key": "recipient",
      "value": "{{sdk.AccAddress of the recipient}}",
      "index": true
    },
    {
      "key": "sender",
      "value": "{{sdk.AccAddress of the sender}}",
      "index": true
    },
    {
      "key": "amount",
      "value": "{{sdk.Coins being transferred}}",
      "index": true
    }
  ]
}
```

```json
{
  "type": "coin_spent",
  "attributes": [
    {
      "key": "spender",
      "value": "{{sdk.AccAddress of the address which is spending coins}}",
      "index": true
    },
    {
      "key": "amount",
      "value": "{{sdk.Coins being spent}}",
      "index": true
    }
  ]
}
```

```json
{
  "type": "coin_received",
  "attributes": [
    {
      "key": "receiver",
      "value": "{{sdk.AccAddress of the address beneficiary of the coins}}",
      "index": true
    },
    {
      "key": "amount",
      "value": "{{sdk.Coins being received}}",
      "index": true
    }
  ]
}
```

#### MintCoins

```json
{
  "type": "coinbase",
  "attributes": [
    {
      "key": "minter",
      "value": "{{sdk.AccAddress of the module minting coins}}",
      "index": true
    },
    {
      "key": "amount",
      "value": "{{sdk.Coins being minted}}",
      "index": true
    }
  ]
}
```

```json
{
  "type": "coin_received",
  "attributes": [
    {
      "key": "receiver",
      "value": "{{sdk.AccAddress of the module minting coins}}",
      "index": true
    },
    {
      "key": "amount",
      "value": "{{sdk.Coins being received}}",
      "index": true
    }
  ]
}
```

#### BurnCoins

```json
{
  "type": "burn",
  "attributes": [
    {
      "key": "burner",
      "value": "{{sdk.AccAddress of the module burning coins}}",
      "index": true
    },
    {
      "key": "amount",
      "value": "{{sdk.Coins being burned}}",
      "index": true
    }
  ]
}
```

```json
{
  "type": "coin_spent",
  "attributes": [
    {
      "key": "spender",
      "value": "{{sdk.AccAddress of the module burning coins}}",
      "index": true
    },
    {
      "key": "amount",
      "value": "{{sdk.Coins being burned}}",
      "index": true
    }
  ]
}
```

## Client

### gRPC

A user can query the precisebank module using gRPC endpoints.

#### TotalFractionalBalances

The `TotalFractionalBalances` endpoint allows users to query the aggregate sum
of all fractional balances. This is primarily used for external verification of
the module state against the reserve balance.

```shell
kava.precisebank.v1.Query/TotalFractionalBalances
```

Example:

```shell
grpcurl -plaintext \
  localhost:9090 \
  kava.precisebank.v1.Query/TotalFractionalBalances
```

Example Output:

```json
{
  "total": "2000000000000akava"
}
```

#### Remainder

The `Remainder` endpoint allows users to query the current remainder amount.

```shell
kava.precisebank.v1.Query/Remainder
```

Example:

```shell
grpcurl -plaintext \
  localhost:9090 \
  kava.precisebank.v1.Query/Remainder
```

Example Output:

```json
{
  "remainder": "100akava"
}
```

#### FractionalBalance

The `FractionalBalance` endpoint allows users to query the fractional balance of
a specific account.

```shell
kava.precisebank.v1.Query/FractionalBalance
```

Example:

```shell
grpcurl -plaintext \
  -d '{"address": "kava1..."}' \
  localhost:9090 \
  kava.precisebank.v1.Query/FractionalBalance
```

Example Output:

```json
{
  "fractional_balance": "10000akava"
}
```
