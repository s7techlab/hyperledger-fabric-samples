package balance

import (
	"github.com/s7techlab/cckit/router"
)

type (
	Store interface {
		Get(router.Context, *BalanceId) (*Balance, error)
		List(router.Context, *BalanceId) ([]*Balance, error)
		Transfer(router.Context, *TransferOperation) error
		TransferBatch(router.Context, []*TransferOperation) error
		Mint(router.Context, *BalanceOperation) error
		Burn(router.Context, *BalanceOperation) error
	}
)
