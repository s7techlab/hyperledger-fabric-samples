package main

import (
	"fmt"

	"github.com/hyperledger/fabric-chaincode-go/shim"

	"github.com/s7techlab/hyperledger-fabric-samples/samples/cpaper/chaincode"
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
