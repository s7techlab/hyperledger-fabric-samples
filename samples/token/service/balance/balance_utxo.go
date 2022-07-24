package balance

import (
	"strings"

	"github.com/s7techlab/cckit/router"
)

var _ Store = &UTXOStore{}

type UTXOStore struct {
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

	senderOutputs, err := u.ListOutputs(ctx, &BalanceId{
		Address: transfer.Sender,
		Symbol:  transfer.Symbol,
		Group:   transfer.Group,
	})
	if err != nil {
		return err
	}

	txID := ctx.Stub().GetTxID()
	useOutputIds, outputsAmount, err := selectOutputsForAmount(senderOutputs, transfer.Amount)
	if err != nil {
		return err
	}

	recipientOutput := &UTXO{
		Address: transfer.Recipient,
		Symbol:  transfer.Symbol,
		Group:   strings.Join(transfer.Group, `,`),
		TxId:    txID,
		Inputs:  useOutputIds,
		Amount:  transfer.Amount,
		//Meta: transfer.Meta,
	}

	if err := State(ctx).Insert(recipientOutput); err != nil {
		return err
	}

	if outputsAmount > transfer.Amount {
		senderChangeOutput := &UTXO{
			Address: transfer.Sender,
			Symbol:  transfer.Symbol,
			Group:   strings.Join(transfer.Group, `,`),
			TxId:    txID,
			Inputs:  useOutputIds,
			Amount:  outputsAmount - transfer.Amount,
		}

		if err := State(ctx).Insert(senderChangeOutput); err != nil {
			return err
		}
	}

	for _, outputId := range useOutputIds {
		utxoId := &UTXOId{
			Address: transfer.Sender,
			Symbol:  transfer.Symbol,
			Group:   strings.Join(transfer.Group, `,`),
			TxId:    outputId,
		}
		if err := State(ctx).Delete(utxoId); err != nil {
			return err
		}
	}

	return nil
}

func (u *UTXOStore) TransferBatch(ctx router.Context, requests []*TransferOperation) error {
	//TODO implement me
	panic("implement me")
}

func selectOutputsForAmount(outputs []*UTXO, amount uint64) ([]string, uint64, error) {
	var (
		selectedOutputsTxId []string
		curAmount           uint64
	)

	for _, o := range outputs {
		selectedOutputsTxId = append(selectedOutputsTxId, o.TxId)
		curAmount += o.Amount
		if curAmount >= amount {
			return selectedOutputsTxId, curAmount, nil
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

func (u *UTXOStore) Burn(context router.Context, operation *BalanceOperation) error {
	//TODO implement me
	panic("implement me")
}
