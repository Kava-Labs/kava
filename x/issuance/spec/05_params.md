
<!--
order: 5
-->

# Parameters

The issuance module has the following parameters:

| Key        | Type           | Example       | Description                                 |
|------------|----------------|---------------|---------------------------------------------|
| Assets     | array (Asset)  | [{see below}] | array of assets created via issuance module |


Each `Asset` has the following parameters

| Key               | Type                   | Example                                         | Description                                           |
|-------------------|------------------------|-------------------------------------------------|-------------------------------------------------------|
| Owner             | sdk.AccAddress         | "kava1cd8z53n7gh2hvz0lmmkzxkysfp5pghufat3h4a"   | the address that controls the issuance of the asset   |
| BlockedAccounts   | array (sdk.AccAddress) | ["kava1tp9u8t8ang53a8tjh2mhqvvwdngqzjvmp3mamc"] | addresses which are blocked from holding the asset    |
| Paused            | boolean                | false                                           | boolean for if issuance and redemption are paused     |
