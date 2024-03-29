package balance

import (
	"fmt"
	"strings"

	"github.com/s7techlab/cckit/router"
	"github.com/s7techlab/cckit/state"
)

var _ Store = &AccountStore{}

type AccountStore struct {
}

func NewAccountStore() *AccountStore {
	return &AccountStore{}
}

func (s *AccountStore) Get(ctx router.Context, id *BalanceId) (*Balance, error) {
	if err := router.ValidateRequest(id); err != nil {
		return nil, err
	}
	balance, err := State(ctx).Get(id, &Balance{})
	if err != nil {
		if strings.Contains(err.Error(), state.ErrKeyNotFound.Error()) {
			// default zero balance even if no Balance state entry for account exists
			return &Balance{
				Address: id.Address,
				Symbol:  id.Symbol,
				Group:   id.Group,
				Amount:  0,
			}, nil
		}
		return nil, err
	}

	return balance.(*Balance), nil
}

func (s *AccountStore) List(ctx router.Context, id *BalanceId) ([]*Balance, error) {
	balances, err := State(ctx).List(&Balance{})
	if err != nil {
		return nil, err
	}
	return balances.(*Balances).Items, nil
}

func (s *AccountStore) Transfer(ctx router.Context, transfer *TransferOperation) error {
	if err := router.ValidateRequest(transfer); err != nil {
		return err
	}
	// subtract from sender balance
	if _, err := s.sub(ctx, &BalanceOperation{
		Address: transfer.Sender,
		Symbol:  transfer.Symbol,
		Group:   transfer.Group,
		Amount:  transfer.Amount,
	}); err != nil {
		return fmt.Errorf(`subtract from sender: %w`, err)
	}
	// add to recipient balance
	if _, err := s.add(ctx, &BalanceOperation{
		Address: transfer.Recipient,
		Symbol:  transfer.Symbol,
		Group:   transfer.Group,
		Amount:  transfer.Amount,
	}); err != nil {
		return fmt.Errorf(`add to recipient: %w`, err)
	}
	return nil
}

func (s *AccountStore) TransferBatch(ctx router.Context, transfers []*TransferOperation) error {
	// todo: COUNT TOTAL AMOUNT !!!
	for _, t := range transfers {
		if err := s.Transfer(ctx, t); err != nil {
			return err
		}
	}

	return nil
}

func (s *AccountStore) Mint(ctx router.Context, op *BalanceOperation) error {
	if err := router.ValidateRequest(op); err != nil {
		return err
	}
	// add to recipient balance
	if _, err := s.add(ctx, op); err != nil {
		return fmt.Errorf(`add: %w`, err)
	}

	return nil
}

func (s *AccountStore) Burn(ctx router.Context, op *BalanceOperation) error {
	if err := router.ValidateRequest(op); err != nil {
		return err
	}
	if _, err := s.sub(ctx, op); err != nil {
		return fmt.Errorf(`sub: %w`, err)
	}

	return nil
}

func (s *AccountStore) add(ctx router.Context, op *BalanceOperation) (*Balance, error) {
	balance, err := s.Get(ctx, &BalanceId{Address: op.Address, Symbol: op.Symbol, Group: op.Group})
	if err != nil {
		return nil, err
	}

	newBalance := &Balance{
		Address: op.Address,
		Symbol:  op.Symbol,
		Group:   op.Group,
		Amount:  balance.Amount + op.Amount,
	}

	if err = State(ctx).Put(newBalance); err != nil {
		return newBalance, err
	}
	return newBalance, err
}

func (s *AccountStore) sub(ctx router.Context, op *BalanceOperation) (*Balance, error) {
	balance, err := s.Get(ctx, &BalanceId{Address: op.Address, Symbol: op.Symbol, Group: op.Group})
	if err != nil {
		return nil, err
	}

	if balance.Amount < op.Amount {
		return nil, fmt.Errorf(`subtract from=%s: %w`, op.Address, ErrAmountInsuficcient)
	}

	newBalance := &Balance{
		Address: op.Address,
		Symbol:  op.Symbol,
		Group:   op.Group,
		Amount:  balance.Amount - op.Amount,
	}

	if err = State(ctx).Put(newBalance); err != nil {
		return newBalance, err
	}
	return newBalance, err
}
