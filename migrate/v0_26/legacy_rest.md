# Legacy REST Endpoint Removal

Kava 16 (v0.26.x) upgrades cosmos-sdk to v0.47.10+. Legacy REST endpoints, which were previously marked as deprecated, are now [removed](https://github.com/cosmos/cosmos-sdk/blob/v0.47.10/UPGRADING.md#appmodule-interface). 

All consumers of these endpoints must migrate to the GRPC generated REST endpoints. See [Kava Chain Swagger](https://swagger.kava.io/?urls.primaryName=Edge ) for supported REST API Methods. 