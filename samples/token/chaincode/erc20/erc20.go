package erc20

import (
	"errors"

	"github.com/s7techlab/cckit/gateway"
	"github.com/s7techlab/cckit/router"

	"github.com/s7techlab/hyperledger-fabric-samples/samples/token/service/account"
	"github.com/s7techlab/hyperledger-fabric-samples/samples/token/service/allowance"
	"github.com/s7techlab/hyperledger-fabric-samples/samples/token/service/balance"
	"github.com/s7techlab/hyperledger-fabric-samples/samples/token/service/burnable"
	"github.com/s7techlab/hyperledger-fabric-samples/samples/token/service/config"
	"github.com/s7techlab/hyperledger-fabric-samples/samples/token/service/config_erc20"
)

var (
	Symbol = `@`

	// 	Token  Static hardcoded token
	Token = &config.CreateTokenTypeRequest{
		Name:        `SomeToken`,
		Symbol:      Symbol,
		Decimals:    2,
		TotalSupply: 10000000,
	}
)

// Gateways for communicating with chaincode
func Gateways(instance gateway.ChaincodeInstance) []gateway.Service {
	gateways := []gateway.Service{
		config_erc20.NewConfigERC20ServiceGatewayFromInstance(instance).ServiceDef(),
		account.NewAccountServiceGatewayFromInstance(instance).ServiceDef(),
		balance.NewBalanceServiceGatewayFromInstance(instance).ServiceDef(),
		allowance.NewAllowanceServiceGatewayFromInstance(instance).ServiceDef(),
		burnable.NewBurnableServiceGatewayFromInstance(instance).ServiceDef(),
	}

	return gateways
}

func New(name string, store balance.Store) (*router.Chaincode, error) {
	r := router.New(name)

	// accountSvc resolves address as base58( invoker.Cert.PublicKey )
	accountSvc := account.NewLocalService()
	configSvc := config.NewStateService()
	// Balance management service with Account storage model

	balanceSvc := balance.New(accountSvc, configSvc, store)
	// Allowance management service
	allowanceSvc := allowance.NewService(accountSvc, store)

	burnableSvc := burnable.NewService(accountSvc, store)
	erc20ConfigSvc := &config_erc20.ERC20Service{Token: configSvc}

	r.Init(func(ctx router.Context) (interface{}, error) {
		// add token definition to state if not exists
		tokenId, err := config.CreateDefaultToken(ctx, configSvc, Token)
		if err != nil {
			if errors.Is(err, config.ErrTokenAlreadyExists) {
				return nil, nil
			}
			return nil, err
		}

		// get chaincode instantiator address
		ownerAddress, err := accountSvc.GetInvokerAddress(ctx, nil)
		if err != nil {
			return nil, err
		}

		// add  `TotalSupply` to chaincode first committer
		if err = balanceSvc.Store.Mint(ctx, &balance.BalanceOperation{
			Address: ownerAddress.Address,
			Symbol:  tokenId.Symbol,
			Group:   tokenId.Group,
			Amount:  Token.TotalSupply,
			Meta:    nil,
		}); err != nil {
			return nil, err
		}

		return nil, nil
	})

	if err := balance.RegisterBalanceServiceChaincode(r, balanceSvc); err != nil {
		return nil, err
	}
	if err := account.RegisterAccountServiceChaincode(r, accountSvc); err != nil {
		return nil, err
	}
	if err := config_erc20.RegisterConfigERC20ServiceChaincode(r, erc20ConfigSvc); err != nil {
		return nil, err
	}
	if err := allowance.RegisterAllowanceServiceChaincode(r, allowanceSvc); err != nil {
		return nil, err
	}
	if err := burnable.RegisterBurnableServiceChaincode(r, burnableSvc); err != nil {
		return nil, err
	}

	return router.NewChaincode(r), nil
}
