0) Create the genesis state for the blockchain by running the following from the `swagger-ui` directory:

`./startchain.sh`

1) Build the hooks file by running the following from the `swagger-ui` folder:

`rm hooks`
`go build hooks.go`

2) Then start the chain:

`kvd unsafe-reset-all`
`kvd start`

3) Then start the `REST` server:

`kvcli rest-server --laddr tcp://127.0.0.1:1317`

4) Then run the following command from the `swagger-ui` folder to run the `dredd` tests:
**IMPORTANT NOTE**: Sometimes the first time you run the dredd tests with hooks the transactions may not have posted to the blockchain so some tests may erroneously fail. When you run the script a second time 
they should pass.

`dredd ../swagger-ui/swagger.yaml localhost:1317 --hookfiles=hooks --language=go --loglevel=debug`

If you run into `dredd` timeout issues you can try:
`dredd swagger.yaml localhost:1317 --hookfiles=hooks --language=go --loglevel=debug --hooks-worker-connect-timeout="3000000"`

**IMPORTANT NOTE**: If you start getting `Validator set is different` errors then you need to try starting the chain from scratch (do NOT just use `unsafe-reset-all`, instead use `./startchain.sh`)