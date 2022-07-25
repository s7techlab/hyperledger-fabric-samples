package main

import (
	"fmt"

	"github.com/hyperledger/fabric-chaincode-go/shim"

	"github.com/s7techlab/hyperledger-fabric-samples/samples/token/chaincode/erc20"
	"github.com/s7techlab/hyperledger-fabric-samples/samples/token/service/balance"
)

func main() {
	cc, err := erc20.New(`erc-20 account`, balance.NewAccountStore())
	if err != nil {
		fmt.Printf("Error creating chaincode: %s", err)
	}

	err = shim.Start(cc)
	if err != nil {
		fmt.Printf("Error starting chaincode: %s", err)
	}
}
