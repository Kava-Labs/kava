# `x/precisebank`

## Abstract

This document specifies the precisebank module of Kava.

The precisebank module is responsible for extending the precision of `x/bank`, intended to be used for the `x/evm`. It serves as a wrapper of `x/bank` to increase the precision of KAVA from 6 to 18 decimals, while preserving the behavior of existing `x/bank` balances.

This module is used only by `x/evm` where 18 decimal points are expected.

## Background

The standard unit of currency on the Kava Chain is `KAVA`.  This is denominated by the atomic unit `ukava`, which represents $10^{-6}$ `KAVA` and there are $10^6$ `ukava` per `KAVA`.

In order to support 18 decimals of precision while maintaining `ukava` as the cosmos native atomic unit, we further split each `ukava` unit into $10^{12}$ `akava` units, the native currency of the Kava EVM.

This gives a full $10^{18}$ precision on the EVM. In order to avoid confusion with atomic `ukava` units, we will refer to `akava` as sub-atomic units.

To review we have:
 - `ukava`, the cosmos native unit and atomic unit of the Kava chain
 - `akava`, the evm native unit and sub-atomic unit of the Kava chain

In order to maintain consistency between the `akava` supply and the `ukava` supply, we add the constraint that each `akava` or sub-atomic unit, may only exist as part of a atomic `ukava`, or in other words, we require each `akava` to be fully backed by a `ukava` in the `x/bank` module.

This is a requirement since `x/bank` `ukava` balances are shared between the cosmos modules and the EVM.  We are wrapping and extending the `x/bank` module with the `x/precisebank` module to add an extra $10^{12}$ units of precision.  If $10^{12}$ `akava` is transferred in the EVM, the cosmos modules will see a 1 `ukava` transfer and vice versa.  If `akava` was not fully backed by `ukava`, then balance changes would not be fully consistent across the cosmos and the EVM.

This brings us to how are account balances are extended to represent `akava` balances larger than $10^{12}$.  First we define $a(n)$, $b(n)$, and $C$ where $a(n)$ is the `akava` balance of account `n`, $b(n)$ is the `ukava` balance of account `n` stored in the `x/bank` module, and $C$ is the conversion factor equal to $10^{12}$.

Any $a(n)$ divisible by $C$, can be represented by $C$ * $b(n)$.  Any remainder not divisible by $C$, we define as $f(n)$ and store this in the `x/precisebank` store.

Thus,

$$a(n) = b(n) \cdot C + f(n)$$

where

$$0 \le f(n) < C$$

$$a(n), b(n) \ge 0$$

This is the quotient-remainder theorem and any $a(n)$ can be represented by unique integers $b(n)$, $f(n)$ where

$$b(n) = \lfloor a(n)/C \rfloor$$

$$f(n) = a(n)\bmod{C}$$

With this definition in mind we will refer to $b(n)$ units as integer units, and $f(n)$ as fractional units.

Now since $f(n)$ is stored in the `x/precisebank` and not tracked by the `x/bank` keeper, these are not counted in the `ukava` supply, so we define

$$T_a = \sum_{n \in \mathcal{A}}{a(n)}$$

$$T_b = \sum_{n \in \mathcal{A}}{b(n)}$$

where $\mathcal{A}$ is the set of all accounts and $T_a$ is the total `akava` supply and $T_b$ is the total `ukava` supply, then a reserve account is added such that

$$a(R) = 0$$

$$b(R) \cdot C = \sum_{n \in \mathcal{A}}{f(n)} + r$$

where R is the reserve account or module account of the `x/precisebank`, and $r$ is the remainder or fractional amount backed by $b(R)$, but not yet in circulation such that

$$T_a = T_b \cdot C - r$$

and

$$ 0 <= r < C$$

We see that $0 <= T_b \cdot C - T_a < C$, and if we mint, burn, or transfer `akava` such that this inequality would be invalid after updates to account balances, we adjust the $T_b$ supply by minting or burning to a reserve account which holds `ukava` equal to that of all `akava` balances less than `C` plus the remaining `akava` balance not in circulation.

If we didn't add these constraints, then the total supply of `ukava` reported by the bank keeper would not account for the `akava` units.  We could increase the supply of `akava` without increasing the reported total supply of KAVA.

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

With these two formulas, we can determine the new remainder and reserve.

$$(b'(R)-b(R)) \cdot C = \sum_{n \in \mathcal{A}}{f'(n)} - \sum_{n \in \mathcal{A}}{f(n)} + r' - r$$

Since only two accounts are involved in the transfer, the total supply of fractional units is updated as

$$(b'(R)-b(R)) \cdot C = f'(1) - f(1) + f'(2) - f(2) + r' - r$$

#### Remainder does not change

$$(b'(R)-b(R)) \cdot C = f'(1) - f(1) + f'(2) - f(2) + r' - r \mod{C}$$

$$0 = (f(1) - a)\bmod{C} - f(1) + (f(2) + a)\bmod{C} - f(2) + r' - r \mod{C}$$

$$0 = f(1) - a - f(1) + f(2) + a - f(2) + r' - r \mod{C}$$

$$0 = r' - r \mod{C}$$

By the quotient remainder theorem, we can rewrite it with the following formula.

$$q * C = r' - r$$

$$0 \leq r' < C, 0 \leq r < C$$

$$ -C < r' - r < C$$

$$q = 0$$

$$ r' - r = 0$$

#### Reserve

$$(b'(R)-b(R)) \cdot C = f'(1) - f(1) + f'(2) - f(2)$$

$$f'(1) = f(1) - a \mod{C}$$

$$f'(1)\bmod{C}= f(1)\bmod{C} - a \bmod{C} \mod{C}$$

$$f'(1) = f(1) - a \bmod{C} \mod{C}$$

$$f'(1) = \begin{cases} f(1) - a\bmod{C} & f'(1) \leq f(1) \\
f(1) - a\bmod{C} + C & f'(1) > f(1) \end{cases}$$

$$f'(2) = f(2) + a \mod{C}$$

$$f'(2)\bmod{C}= f(2)\bmod{C} + a \bmod{C} \mod{C}$$

$$f'(2) = f(2) + a \bmod{C} \mod{C}$$

$$f'(2) = \begin{cases} f(2) + a\bmod{C} & f'(2) \geq f(2) \\
f(2) + a\bmod{C} - C & f'(2) < f(2) \end{cases}$$

$$b'(R) - b(R) \cdot C = \begin{cases} f(1) - a\bmod{C} + C - f(1) + f(2) + a\bmod{C} - C + f(2) & f'(1) > f(1) \land f'(2) < f(2) \\
f(1) - a\bmod{C} - f(1) + f(2) + a\bmod{C} - C + f(2) & f'(1) \leq f(1) \land f'(2) < f(2) \\
f(1) - a\bmod{C} + C - f(1) + f(2) + a\bmod{C} + f(2) & f'(1) > f(1) \land f'(2) \geq f(2) \\
f(1) - a\bmod{C} - f(1) + f(2) + a\bmod{C} + f(2) & f'(1) \leq f(1) \land f'(2) \geq f(2) \\
\end{cases}$$

$$b'(R) - b(R) \cdot C = \begin{cases} 0 & f'(1) > f(1) \land f'(2) < f(2) \\
-C & f'(1) \leq f(1) \land f'(2) < f(2) \\
C & f'(1) > f(1) \land f'(2) \geq f(2) \\
0 & f'(1) \leq f(1) \land f'(2) \geq f(2) \\
\end{cases}$$

$$b'(R) - b(R) = \begin{cases} 0 & f'(1) > f(1) \land f'(2) < f(2) \\
-1 & f'(1) \leq f(1) \land f'(2) < f(2) \\
1 & f'(1) > f(1) \land f'(2) \geq f(2) \\
0 & f'(1) \leq f(1) \land f'(2) \geq f(2) \\
\end{cases}$$

### Burn

$$a'(1) = a(1) - a$$

$$(b'(R)-b(R)) \cdot C = f'(1) - f(1) + r' - r$$

$$f'(1) = f(1) - a \mod{C}$$

$$f'(1)\bmod{C}= f(1)\bmod{C} - a \bmod{C} \mod{C}$$

$$f'(1) = f(1) - a \bmod{C} \mod{C}$$

$$f'(1) = \begin{cases} f(1) - a\bmod{C} & f'(1) \leq f(1) \\
f(1) - a\bmod{C} + C & f'(1) > f(1) \end{cases}$$

$$r' = r + a \mod{C}$$

$$r'\bmod{C}= r\bmod{C} + a \bmod{C} \mod{C}$$

$$r' = r + a \bmod{C} \mod{C}$$

$$r' = \begin{cases} r + a\bmod{C} & r' \geq r \\
r + a\bmod{C} - C & r' < r \end{cases}$$

$$b'(R) - b(R) = \begin{cases} 0 & f'(1) > f(1) \land r' < r \\
-1 & f'(1) \leq f(1) \land r' < r \\
1 & f'(1) > f(1) \land r' \geq r \\
0 & f'(1) \leq f(1) \land r' \geq r \\
\end{cases}$$

### Mint

$$a'(1) = a(1) + a$$

$$(b'(R)-b(R)) \cdot C = f'(1) - f(1) + r' - r$$

$$f'(1) = f(1) + a \mod{C}$$

$$f'(1)\bmod{C}= f(1)\bmod{C} + a \bmod{C} \mod{C}$$

$$f'(1) = f(1) + a \bmod{C} \mod{C}$$

$$f'(1) = \begin{cases} f(1) + a\bmod{C} & f'(1) \geq f(1) \\
f(1) + a\bmod{C} - C & f'(1) < f(1) \end{cases}$$

$$r' = r - a \mod{C}$$

$$r'\bmod{C}= r\bmod{C} - a \bmod{C} \mod{C}$$

$$r' = r - a \bmod{C} \mod{C}$$

$$r' = \begin{cases} r - a\bmod{C} & r' \leq r \\
r - a\bmod{C} + C & r' > r \end{cases}$$

$$b'(R) - b(R) = \begin{cases} 0 & r' > r \land f'(1) < f(1) \\
-1 & r' \leq r \land f'(1) < f(1) \\
1 & r' > r \land f'(1) \geq f(1) \\
0 & r' \leq r \land f'(1) \geq f(1) \\
\end{cases}$$
