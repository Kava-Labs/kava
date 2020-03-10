1) Build the hooks file by running the following from the `testthing` folder:

`rm testthing`
`go build`

2) Then start the chain:

`kvd unsafe-reset-all`
`kvd start`

3) Then start the `REST` server:

`kvcli rest-server --laddr tcp://127.0.0.1:1317`

4) Then run the following command from the `testthing` folder to run the `dredd` tests:

`dredd ../swagger-ui/swagger.yaml localhost:1317 --hookfiles=testthing --language=go --loglevel=debug`