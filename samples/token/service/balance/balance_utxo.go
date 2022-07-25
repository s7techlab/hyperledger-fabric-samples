package balance

import (
	"errors"
	"strings"

	"github.com/s7techlab/cckit/router"
)

var _ Store = &UTXOStore{}

var (
	ErrSenderRecipientEqual = errors.New(`sender recipient equal`)
	ErrSenderNotEqual       = errors.New(`sender not equal`)
	ErrSymbolNotEqual       = errors.New(`symbol not equal`)
	ErrRecipientDuplicate   = errors.New(`errors recipient duplicate`)
)

type UTXOStore struct {
}

func (u *UTXO) ID() *UTXOId {
	return &UTXOId{
		Address: u.Address,
		Symbol:  u.Symbol,
		Group:   u.Group,
		TxId:    u.TxId,
	}
}

func NewUTXOStore() *UTXOStore {
	return &UTXOStore{}
}

func (u *UTXOStore) Get(ctx router.Context, balanceId *BalanceId) (*Balance, error) {
	if err := router.ValidateRequest(balanceId); err != nil {
		return nil, err
	}
	outputs, err := u.ListOutputs(ctx, balanceId)
	if err != nil {
		return nil, err
	}
	var amount uint64
	for _, u := range outputs {
		amount += u.Amount
	}

	balance := &Balance{
		Address: balanceId.Address,
		Symbol:  balanceId.Symbol,
		Group:   balanceId.Group,
		Amount:  amount,
	}

	return balance, nil
}

//ListOutputs unspended outputs list
func (u *UTXOStore) ListOutputs(ctx router.Context, balanceId *BalanceId) ([]*UTXO, error) {
	utxos, err := State(ctx).ListWith(&UTXO{}, UTXOKeyBase(&UTXO{
		Address: balanceId.Address,
		Symbol:  balanceId.Symbol,
		Group:   strings.Join(balanceId.Group, `,`),
	}))
	if err != nil {
		return nil, err
	}

	return utxos.(*UTXOs).Items, nil
}

func (u *UTXOStore) List(ctx router.Context, id *BalanceId) ([]*Balance, error) {
	//TODO implement me
	panic("implement me")
}

func (u *UTXOStore) Transfer(ctx router.Context, transfer *TransferOperation) error {
	if err := router.ValidateRequest(transfer); err != nil {
		return err
	}

	if transfer.Sender == transfer.Recipient {
		return ErrSenderRecipientEqual
	}

	senderOutputs, err := u.ListOutputs(ctx, &BalanceId{
		Address: transfer.Sender,
		Symbol:  transfer.Symbol,
		Group:   transfer.Group,
	})
	if err != nil {
		return err
	}

	useOutputs, outputsAmount, err := selectOutputsForAmount(senderOutputs, transfer.Amount)
	if err != nil {
		return err
	}

	txID := ctx.Stub().GetTxID()
	recipientOutput := &UTXO{
		Address: transfer.Recipient,
		Symbol:  transfer.Symbol,
		Group:   strings.Join(transfer.Group, `,`),
		TxId:    txID,
		Inputs:  nil,
		Amount:  transfer.Amount,
		//Meta: transfer.Meta,
	}

	if err := State(ctx).Insert(recipientOutput); err != nil {
		return err
	}

	// chgange output
	if outputsAmount > transfer.Amount {
		senderChangeOutput := &UTXO{
			Address: transfer.Sender,
			Symbol:  transfer.Symbol,
			Group:   strings.Join(transfer.Group, `,`),
			TxId:    txID,
			Inputs:  nil, // do we need this ?
			Amount:  outputsAmount - transfer.Amount,
		}

		if err := State(ctx).Insert(senderChangeOutput); err != nil {
			return err
		}
	}

	for _, output := range useOutputs {
		if err := State(ctx).Delete(output.ID()); err != nil {
			return err
		}
	}

	return nil
}

func (u *UTXOStore) TransferBatch(ctx router.Context, transfers []*TransferOperation) error {
	var (
		sender, symbol string
		group          []string
		recipients     = make(map[string]interface{})
		totalAmount    uint64
	)
	for _, transfer := range transfers {

		if err := router.ValidateRequest(transfer); err != nil {
			return err
		}

		if sender == `` {
			sender = transfer.Sender
		}

		if transfer.Sender != sender {
			return ErrSenderNotEqual
		}

		if sender == transfer.Recipient {
			return ErrSenderRecipientEqual
		}
		if symbol == `` {
			symbol = transfer.Symbol
		}

		if transfer.Symbol != symbol {
			return ErrSymbolNotEqual
		}

		if len(transfer.Group) > 0 {
			panic(`implement me`)
		}

		if _, ok := recipients[transfer.Recipient]; ok {
			return ErrRecipientDuplicate
		}
		recipients[transfer.Recipient] = nil
		totalAmount += transfer.Amount
	}

	senderOutputs, err := u.ListOutputs(ctx, &BalanceId{
		Address: sender,
		Symbol:  symbol,
		Group:   group,
	})
	if err != nil {
		return err
	}

	useOutputs, outputsAmount, err := selectOutputsForAmount(senderOutputs, totalAmount)
	if err != nil {
		return err
	}

	for _, output := range useOutputs {
		if err := State(ctx).Delete(output.ID()); err != nil {
			return err
		}
	}

	txID := ctx.Stub().GetTxID()

	for _, transfer := range transfers {
		recipientOutput := &UTXO{
			Address: transfer.Recipient,
			Symbol:  transfer.Symbol,
			Group:   strings.Join(group, `,`),
			TxId:    txID,
			Inputs:  nil,
			Amount:  transfer.Amount,
			//Meta: transfer.Meta,
		}

		if err := State(ctx).Insert(recipientOutput); err != nil {
			return err
		}
	}

	if outputsAmount > totalAmount {
		senderChangeOutput := &UTXO{
			Address: sender,
			Symbol:  symbol,
			Group:   strings.Join(group, `,`),
			TxId:    txID,
			Inputs:  nil, // do we need this ?
			Amount:  outputsAmount - totalAmount,
		}

		if err := State(ctx).Insert(senderChangeOutput); err != nil {
			return err
		}
	}

	return nil
}

// todo: Optimize selection, to maximum fit outputs
func selectOutputsForAmount(outputs []*UTXO, amount uint64) ([]*UTXO, uint64, error) {
	var (
		selectedOutputs []*UTXO
		curAmount       uint64
	)

	for _, o := range outputs {
		selectedOutputs = append(selectedOutputs, o)
		curAmount += o.Amount
		if curAmount >= amount {
			return selectedOutputs, curAmount, nil
		}
	}

	return nil, 0, ErrAmountInsuficcient
}

func (u *UTXOStore) Mint(ctx router.Context, op *BalanceOperation) error {
	mintedOutput := &UTXO{
		Address: op.Address,
		Symbol:  op.Symbol,
		Group:   strings.Join(op.Group, `,`),
		TxId:    ctx.Stub().GetTxID(),
		Inputs:  nil,
		Amount:  op.Amount,
	}

	return State(ctx).Insert(mintedOutput)
}

func (u *UTXOStore) Burn(ctx router.Context, burn *BalanceOperation) error {
	outputs, err := u.ListOutputs(ctx, &BalanceId{
		Address: burn.Address,
		Symbol:  burn.Symbol,
		Group:   burn.Group,
	})
	if err != nil {
		return err
	}

	useOutputs, outputsAmount, err := selectOutputsForAmount(outputs, burn.Amount)
	if err != nil {
		return err
	}

	for _, output := range useOutputs {
		if err := State(ctx).Delete(output.ID()); err != nil {
			return err
		}
	}

	if outputsAmount > burn.Amount {
		senderChangeOutput := &UTXO{
			Address: burn.Address,
			Symbol:  burn.Symbol,
			Group:   strings.Join(burn.Group, `,`),
			TxId:    ctx.Stub().GetTxID(),
			Inputs:  nil, // do we need this ?
			Amount:  outputsAmount - burn.Amount,
		}

		if err := State(ctx).Insert(senderChangeOutput); err != nil {
			return err
		}
	}

	return nil
}
