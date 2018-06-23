To initialise a network:

 - delete everything (including persistant volume claim)
 - deploy everything except the deployments
 - wait until the job has finished, then deploy the deployments

 Note on config

  - Secrets and configmaps need to be generated from files
  - Ideally everything would be in one file but kubectl doesn't scan directories yet: https://github.com/kubernetes/kubernetes/issues/62421
  - `kubectl create secret generic kava-user-keys --from-file=./init/init-data --dry-run -o yaml > secret-user.yml`
  - `kubectl create secret generic kava-node-config --from-file=./init/init-data/.kvd/config --dry-run -o yaml > secret-config.yml`

Examples of using light client with the node:

 - Get the status `kvcli status --node <node's-url>:46657 --chain-id test-kava`
 - Send coins `kvcli send --name <your-key-name> --to <receiver's-address> --amount 100KVA --node <node's-url>:46657 --chain-id test-kava`
 - Run the light client daemon `kvcli rest-server --node <node's-url>:46657 --chain-id test-kava`
 