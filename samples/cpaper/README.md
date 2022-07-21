# Commercial paper chaincode & API to interact with chaincode example

This is an extended example of the
official [Commercial paper scenario](https://hyperledger-fabric.readthedocs.io/en/release-1.4/developapps/scenario.html)

### Features

* [Chaincode model](cpaper.proto) interfacem transaction payload, event definitions and chaincode state schema
* [Tests](cpaper_test.go) covers service implementation - no need to convert chaincode input payload and response
* [Mocked API](cmd/api-mocked) to interact with mocked chaincode