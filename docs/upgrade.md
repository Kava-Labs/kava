
# Validator Upgrade

These are some guidelines to upgrade to a new testnet if you where validating on a previous one.

 1. Stop the current validator.

 1. Remove the old config.

        rm -r $HOME/.kvd
        rm -r $HOME/.kvcli

 1. Get the latest code.
 
        cd $GOPATH/src/github.com/kava-labs/kava
        git pull
    
 1. Get the latest dependencies.
 
        dep ensure -vendor-only

 1. Install.

        go install ./cmd/kvd
        go install ./cmd/kvcli

 1. Follow the installation instructions for running a full node and becoming a validator.
