#! /bin/bash

# THIS IS A BASH SCRIPT THAT STOPS THE BLOCKCHAIN AND THE REST API
echo "Stopping blockchain"
pgrep kvd | xargs kill
pgrep kvcli | xargs kill
echo "COMPLETED"