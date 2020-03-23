#! /bin/bash

# THIS IS A BASH SCRIPT THAT STOPS THE BLOCKCHAIN AND THE REST API
pgrep kvd | xargs kill
pgrep kvcli | xargs kill