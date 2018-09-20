# Payment Channels

Payment channels are designed to enable high speed and throughput for transactions while requiring no counter-party risk.

This initial implementation is for unidirectional channels. Channels can be opened by a sender and closed immediately by the receiver, or by the sender subject to a dispute period. There are no top-ups or partial withdrawals (yet).


# Usage
>The following commands require communication with a full node. By default they expect one to be running locally (accessible on localhost), but a remote can be provided with the `--node` flag.

## Create a channel

	kvcli paychan create --from <your account name> --to <receivers address> --amount 100KVA

## Send an off-chain payment
Send a payment for 10 KVA.

	kvcli paychan pay --from <your account name> --sen-amt 90KVA --rec-amt 10KVA --chan-id <ID of channel> --filename payment.json

Send the file `payment.json` to your receiver. Then they run the following to verify.

	kvcli paychan verify --filename payment.json

## Close a channel
The receiver can close immediately at any time.

	kvcli paychan submit --from <receiver's account name> --payment payment.json

The sender can submit a close request, causing the channel will close automatically after a dispute period. During this period a receiver can still close immediately, overruling the sender's request.

	kvcli paychan submit --from <receiver's account name> --payment payment.json

>Note: The dispute period on the testnet is 30 seconds for ease of testing.

## Get info on a channel

	kvcli get --chan-id <ID of channel>

This will print out a channel, if it exists, and any submitted close requests.
