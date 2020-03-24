///////////////////////////////////////////////////////////////////////////////////////
///// IF YOU FOLLOW THESE INSTRUCTIONS ALL THE TESTS SHOULD PASS THE FIRST TIME ///////
///////////////////////////////////////////////////////////////////////////////////////

Instructions on how to run the `dredd` tests

1) Make sure that you have the latest versions of `node` and `npm` and `npx` installed. Then install `dredd` globally using the following command from the `swagger-ui` folder:

`npm install dredd --global`

2) Build the test file by running the following from the `debugging_tools` folder:

`rm test`
`go build test.go`

3) Now create the genesis state for the blockchain, start the blockchain, start the rest 
server, send the required transactions to the blockchain by running the following from the `swagger-ui` 
directory. 
**IMPORTANT** - Wait until the script completes before proceeding (takes about 1-2 minutes):

`./startchain.sh`

4) Then run the below `dredd` command from the `swagger-ui` folder to run the `dredd` tests.
**When you run the script 62 tests should pass and zero should fail** 

(See `test-results` in the `swagger-ui` for what the `dredd` output should look like.)

`dredd swagger.yaml localhost:1317`

If you want to run with the debug logs printed out use the following command:

`dredd swagger.yaml localhost:1317 --loglevel=debug`

5) Finally stop the blockchain and rest server using the following script:

`./stopchain.sh`

**DEBUGGING NOTE**: If you start getting `Validator set is different` errors then you need to try starting the chain from scratch (do NOT just use `unsafe-reset-all`, instead use `./stopchain.sh` and then `./startchain.sh`)


