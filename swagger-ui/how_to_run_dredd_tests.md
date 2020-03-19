Instructions on how to run the `dredd` tests

1) Create the genesis state for the blockchain, start the blockchain and the rest server by running the following from the `swagger-ui` directory (`ctrc-c` to stop both the blockchain and the rest server):

`./startchain.sh`

2) In a separate terminal, build the hooks file by running the following from the `swagger-ui` folder:

`rm hooks`
`go build hooks.go`

3) Then run the below `dredd` command **TWICE** from the `swagger-ui` folder to run the `dredd` tests, ignore the output from the first run. The first time the hooks will send transactions to the chain but sometimes the first time you run the dredd tests with hooks the transactions may not have committed to the blockchain so some tests may erroneously fail. **When you run the script a second time all the tests should pass.** 

See `test-results` in the `swagger-ui` for what the `dredd` output should look like.

`dredd swagger.yaml localhost:1317 --hookfiles=hooks --language=go --loglevel=debug --hooks-worker-connect-timeout="3000000"`

You can also run `dredd` with a lower timeout however this is **NOT** recommmended:
`dredd ../swagger-ui/swagger.yaml localhost:1317 --hookfiles=hooks --language=go --loglevel=debug`

**IMPORTANT NOTE**: If you start getting `Validator set is different` errors then you need to try starting the chain from scratch (do NOT just use `unsafe-reset-all`, instead use `./startchain.sh`)