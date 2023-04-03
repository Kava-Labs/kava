This directory contains contract interfaces used by the e2e test suite.

# Prereq
* Abigen: https://geth.ethereum.org/docs/tools/abigen
* Solidity: https://docs.soliditylang.org/en/latest/installing-solidity.html

# Create Contract Interfaces for Go

If you have the compiled ABI, you can skip directly to step 4.

To create new go interfaces to contracts:
1. add the solidity file: `<filename>.sol`
2. decide on a package name. this will be the name of the package you'll import into go (`<pkg-name>`)
3. compile the abi & bin for the contract: `solc -o <pkg-name> --abi --bin <filename>.sol`
  * run from this directory
  * note that `-o` is the output directory. this will generate `<pkg-name>/<filename>.abi`
4. generate the golang interface:
  `abigen --abi=<pkg-name>/<filename>.abi --bin=<pkg-name>/<filename>.bin --pkg=<pkg-name> --out=<pkg-name>/main.go`
5. import and use the contract in Go.

By including the bin, the generated interface will have a `Deploy*` method. If you only need to interact with an existing contract, you can exclude the `--bin` and only an interaction method interface will be generated.

# Resources
* https://geth.ethereum.org/docs/developers/dapp-developer/native-bindings
