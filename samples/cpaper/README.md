# Commercial paper chaincode & API to interact with chaincode example

This is an extended example of the official [Commercial paper scenario](https://hyperledger-fabric.readthedocs.io/en/release-1.4/developapps/scenario.html)

### Features

* [Protobuf](cpaper.proto) transaction payload, event definitions and chaincode state schema
* [Chaincode](cpaper.proto) interface, defined and [implemented](chaincode/chaincode.go) as gRPC service
* [Tests](chaincode/chaincode_test.go) covers service implementation - no need to convert chaincode input payload and response
* [API](api) to interact with chaincode