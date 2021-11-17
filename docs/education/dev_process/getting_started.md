## Getting Started For Developers 

In order for developers to start building modules, they must install the following tools: 

1. [Go 1.6 or higher](https://golang.org/doc/install)
2. [Docker](https://docs.docker.com/get-docker/)

### Go Programming Language 
Go is used to develop modules and gives much needed performance and flexibility to blockchain developers being a language developed and used by some of the largest companies in the world, in their server/networking applications. 

[A Tour Of Go](https://tour.golang.org/welcome/1)

### Docker 
Docker is a software containerization tool used to spin up Kava nodes and shut them down quickly and allow software portability between different operating systems & environments. It is also used to spin up multiple Kava nodes locally and handles basic networking between them with Docker Compose. 

[Docker Overview](https://docs.docker.com/get-started/overview/)

### Set Up bash_profile
Once Go & Docker are installed, update your bash_profile to include the go path and an alias command for one of the tools we will use to handle Kava node interactions
```
export PATH=/usr/local/go/bin:$PATH
export PATH=$HOME/go/bin:$PATH
export GOPATH=$HOME/go
export GO111MODULE=on

alias dkvcli='docker exec -it generated_kavanode_1 kvcli'
```
Make sure to source your bash profile or restart it for the changes to take place. 

## Getting The Kava Repository & Development Tools 

Once you have the core tools installed & set up, its now time to get the following repositories from Github: 
	

- [kava](https://github.com/Kava-Labs/kava)  
    - Main Kava Repo that holds all modules 
- [kvtool](https://github.com/Kava-Labs/kvtool) 
    - Dev tools to interact with a Kava node 

## Set Up a Local Blockchain 

Now that you have set up all the tools & repositories in your local machine its finally time to set up a local blockchain. 

 - Open a terminal and change into the ```kvtool``` directory.
 - Ensure Docker is running.
 - Run ```make install``` in your terminal which will install ```kvtool``` in your machine. 
 - Ensure Docker is running.
 - Run ```kvtool testnet bootstrap``` this command will build against the master branch from the kava project, initialize the Docker containers and finally starts a local chain. 


Now that you have a local chain running, you can start utilizing the ```dkvcli``` that we set up an alias for. If for whatever reason ```dkvcli``` doesn't work, you can try the following: 

 - Open a terminal and change into the ```kvtool``` directory.
 - In the ```kvtool``` directory there should be a directory named ```full_configs``` change into it. 
 - Once at ```full_configs``` directory change into ```generated``` directory.
 - Once you are at ```generated``` run the following command ```docker-compose exec kavanode bash```. 

This should open up a bash terminal inside a docker container that will give you access to the ```kvcli``` command line interface.  You should see something similar to the snippet below after typing ```kvcli help```: 
```
bash-5.0# kvcli
Command line interface for interacting with kvd

Usage:
  kvcli [command]

Available Commands:
  status      Query remote node for status
  config      Create or query an application CLI configuration file
  query       Querying subcommands
  tx          Transactions subcommands

  rest-server Start LCD (light-client daemon), a local REST server

  keys        Add or view local private keys

  version     Print the app version
  help        Help about any command

Flags:
      --chain-id string   Chain ID of tendermint node
  -e, --encoding string   Binary encoding (hex|b64|btc) (default "hex")
  -h, --help              help for kvcli
      --home string       directory for config and data (default "/root/.kvcli")
  -o, --output string     Output format (text|json) (default "text")
      --trace             print out full stack trace on errors

Use "kvcli [command] --help" for more information about a command.
```