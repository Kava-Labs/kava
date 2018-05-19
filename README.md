To initialise a network:

 - delete everything (including persistant volume claim)
 - deploy everything except the deployment
 - wait until the job has finished, then deploy the deployment
 - check job pod logs for the validator account backup phrase :\

Examples of using light client with the node:

 - Get the status `gaiacli status --node <node's-ip-address>:46657 --chain-id kava`
 - Send coins `gaiacli send --name <your-key-name> --to <receiver's-address> --amount 10kavaToken --node <node's-ip-address>:46657 --chain-id kava`
 - Run the rest server `gaiacli rest-server --node <node's-ip-address>:46657 --chain-id kava`
 
Notes

 - There's two persistant volumes, for `.gaiad` and for `.gaiacli`, because their default locations are awkward.