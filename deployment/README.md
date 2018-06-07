To initialise a network:

 - delete everything (including persistant volume claim)
 - deploy everything except the deployments
 - wait until the job has finished, then deploy deployment-d
 - check job pod logs for the validator account backup phrase :(
 - use `kubectl exec` on the deployment-d pod and use gaiacli to recover the validator
 - do the same to add other keys and move tokens around
 - start up lcd pod (looks like only one instance of gaiacli can access the keys DB)

Examples of using light client with the node:

 - Get the status `gaiacli status --node <node's-ip-address>:46657 --chain-id kava`
 - Send coins `gaiacli send --name <your-key-name> --to <receiver's-address> --amount 10kavaToken --node <node's-ip-address>:46657 --chain-id kava`
 - Run the rest server `gaiacli rest-server --node <node's-ip-address>:46657 --chain-id kava`
 
Notes

 - There's two persistant volumes, for `.gaiad` and for `.gaiacli`, because their default locations are awkward.