package main

import (
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/s7techlab/hyperledger-fabric-samples/commercial-paper/chaincode"
)

func main() {
	cc, err := chaincode.New()
	if err != nil {
		fmt.Printf("Error creating chaincode: %s", err)
	}

	err = shim.Start(cc)
	if err != nil {
		fmt.Printf("Error starting chaincode: %s", err)
	}
}
