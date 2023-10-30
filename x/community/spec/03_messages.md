<!--
order: 3
-->

# Messages

## FundCommunityPool

Send coins directly from the sender to the community module account.

The transaction fails if the amount cannot be transferred from the sender to the community module account.

https://github.com/Kava-Labs/kava/blob/1d36429fe34cc5829d636d73b7c34751a925791b/proto/kava/community/v1beta1/tx.proto#L21-L30

## UpdateParams

Update module parameters via gov proposal.

The transaction fails if the message is not submitted through a gov proposal.
The message `authority` must be the x/gov module account address.

https://github.com/Kava-Labs/kava/blob/1d36429fe34cc5829d636d73b7c34751a925791b/proto/kava/community/v1beta1/tx.proto#L35-L44
