## Support 

### Where did my coins go? What is the deal with the hot wallet balance? 
Check the status of the transaction by searching for your kava address on [Mintscan](https://www.mintscan.io/kava).  
  
If you don’t see your funds as expected, these are the most likely reasons:  
- The deputy is running slowly and the transaction is still in progress  
- Check to be sure that asset is visible in the Tokens tab in Trust Wallet. If the asset is not visible, search for it in the Manage tab and toggle it on.  
  
Every cross-chain swap has an auditable record. In times of high network traffic, the deputy might not complete transactions immediately. If you encounter this issue, please wait patiently. If your transaction has not completed or refunded after 24 hours, please contact the Kava team and send us the transaction ID.  
  
To operate cross-chain swaps, Binance must allocate a certain amount of funds for transfers into a hot wallet. As a security measure, the Binance team needs to manually refill the hot wallet from time to time. Otherwise, the rest of the cross-chain swap is completely automated.
### I have unstaked my coins, why are they still pending?
After unstaking KAVA, those coins are locked for 21 days. After that, they are released and can be used.
### Why can't I repay my CDP?
The total amount to close the position is the original loan amount plus accrued interest.  
  
When repaying a CDP, the remaining debt balance cannot be below 10 USDX.  
If you try to repay your original principal only, the transaction will fail if the remaining interest would leave a debt balance of less than 10 USDX.  
  
If you need to purchase additional USDX to fully close out a position:  
- Swap an asset for USDX on Kava Swap
- Purchase USDX on AscendEx  
- Peer-to-peer transaction through [Kava TipBot](https://kavatipbot.com/)  

### I have enough balance, but it says I can't transfer because my balance is too low
Check if your balance contains vested (“locked”) coins or not. If so, you need to wait for the vesting period to complete.
### I got an error during a transaction
"Error during Broadcasting - could not broadcast transaction"  
This error can result from:  
- closing the app  
- a problem with the connection  
- attempting to repeatedly send the same transaction.  
  
Check for the transaction on the balances tab in the Kava app or the tokens tab (if you're using Cosmostation wallet) and see if it was successfully completed. If not, your tokens will still be there.  
  
“Error during Confirming - out of gas”  
- This transaction requires more than the default amount of gas.  
- Note the amount of “gasUsed”  
- Attempt the transaction again, and select click “Advanced” below the fee slider to be able to manually set the gas amount. If you enter an amount greater than “gasUsed,” the transaction should complete.

### Why are my transactions for 10 KAVA and 10 BNB showing up as 10,000,000 ukava and 1,000,000,000, bnb in Trust Wallet? 
This is how the coins are represented in the source code. They represent the smallest indivisible unit of the currency.  
‍  
- ukava is one -millionth (or 10^-6) of a KAVA coin  
- bnb is one hundred-millionth (or 10^-8) of a BNB coin

### My validator node is having problems
Please contact our team through [Discord](https://discord.com/invite/kQzh3Uv), [Telegram](https://t.me/kavalabs), or Slack. Describe the problem in full with steps to reproduce.  
  
Check the #validator-announcements channel in Discord for information regarding the potential need for an update.  
  
Feel free to message us and we will set up a private channel on Telegram or Slack.
### Somebody from Kava support messaged me in Telegram. Are they legitimate team members?
Kava team members never message first. Please report at @notoscam and block the account. The scammer might also imitate the account of a team member changing the letters in the username. If you need assistance, find the team member in the chat user list and send them a direct message.