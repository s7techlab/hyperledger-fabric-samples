package burnable

import (
	"fmt"

	"github.com/s7techlab/cckit/extensions/owner"
	"github.com/s7techlab/cckit/router"

	"github.com/s7techlab/hyperledger-fabric-samples/samples/token/service/account"
	"github.com/s7techlab/hyperledger-fabric-samples/samples/token/service/balance"
)

type Burnable struct {
	account account.Getter
	balance balance.Store
}

func NewService(account account.Getter, balance balance.Store) *Burnable {
	return &Burnable{
		account: account,
		balance: balance,
	}
}

func (b *Burnable) Burn(ctx router.Context, burn *BurnRequest) (*BurnResponse, error) {
	if err := router.ValidateRequest(burn); err != nil {
		return nil, err
	}

	invokerAddress, err := b.account.GetInvokerAddress(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf(`get invoker address: %w`, err)
	}

	if invokerAddress.Address != burn.Address {
		return nil, owner.ErrOwnerOnly
	}

	if err = b.balance.Burn(ctx, &balance.BalanceOperation{
		Address: invokerAddress.Address,
		Symbol:  burn.Symbol,
		Group:   burn.Group,
		Amount:  burn.Amount,
		Meta:    nil, //todo:
	}); err != nil {
		return nil, err
	}

	if err = Event(ctx).Set(&Burned{
		Address: invokerAddress.Address,
		Symbol:  burn.Symbol,
		Group:   burn.Group,
		Amount:  burn.Amount,
	}); err != nil {
		return nil, err
	}

	return &BurnResponse{
		Address: invokerAddress.Address,
		Symbol:  burn.Symbol,
		Group:   burn.Group,
		Amount:  burn.Amount,
	}, nil
}
