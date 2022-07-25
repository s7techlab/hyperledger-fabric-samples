package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/s7techlab/cckit/identity/testdata"
	"google.golang.org/grpc"

	"github.com/s7techlab/cckit/gateway"
	"github.com/s7techlab/cckit/testing"

	"github.com/s7techlab/hyperledger-fabric-samples/samples/token/chaincode/erc20"
	"github.com/s7techlab/hyperledger-fabric-samples/samples/token/service/balance"
)

const (
	chaincodeName = `erc20`
	channelName   = `erc20`

	grpcAddress = `:8080`
	restAddress = `:8081`
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// default identity for signing requests to peeer (mocked)
	apiIdentity := testdata.Certificates[0].MustIdentity(testdata.DefaultMSP)

	// Create mock for commercial paper chaincode invocation
	// Commercial paper chaincode instance
	cc, err := erc20.New(`erc-20`, balance.NewUTXOStore())
	if err != nil {
		log.Fatalln(err)
	}

	// Mockstub for erc20 chaincode
	erc20Mock := testing.NewMockStub(chaincodeName, cc)
	erc20Mock.From(apiIdentity).Init()

	// Chaincode invocation service mock. For real network you can use example with hlf-sdk-go
	peer := testing.NewPeer().WithChannel(channelName, erc20Mock)

	ccInstance := gateway.NewChaincodeInstanceService(peer, &gateway.ChaincodeLocator{
		Channel:   channelName,
		Chaincode: chaincodeName,
	}, gateway.WithDefaultSigner(apiIdentity))

	// Generated gateway for access to chaincode from external application
	gateways := erc20.Gateways(ccInstance)

	grpcListener, err := net.Listen("tcp", grpcAddress)
	if err != nil {
		log.Fatalf("failed to listen grpc: %v", err)
	}

	// Create gRPC server with services used in chaincode
	s := grpc.NewServer()
	for _, gw := range gateways {
		s.RegisterService(gw.GRPCDesc(), gw.Impl())
	}

	// Runs gRPC server in goroutine
	go func() {
		log.Printf(`listen gRPC at %s`, grpcAddress)
		if err := s.Serve(grpcListener); err != nil {
			log.Fatalf("failed to serve gRPC: %v", err)
		}
	}()

	// wait for gRPC service stared
	time.Sleep(3 * time.Second)

	// Init grpc gateways for REST API
	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithInsecure()}

	for _, gw := range gateways {
		if err := gw.GRPCGatewayRegister()(ctx, mux, grpcAddress, opts); err != nil {
			log.Fatalf("failed to register handler from endpoint %v", err)
		}
	}
	log.Printf(`listen REST at %s`, restAddress)

	// Start HTTP server (and proxy calls to gRPC server endpoint)
	if err = http.ListenAndServe(restAddress, mux); err != nil {
		log.Fatalf("failed to serve REST: %v", err)
	}
}
