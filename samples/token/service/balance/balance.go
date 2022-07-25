package balance

import (
	"fmt"

	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/s7techlab/hyperledger-fabric-samples/samples/token/service/account"
	"github.com/s7techlab/hyperledger-fabric-samples/samples/token/service/config"

	"github.com/s7techlab/cckit/router"
)

type Service struct {
	Account account.Getter
	Token   config.TokenGetter
	Store   Store
}

func New(
	accountResolver account.Getter,
	tokenGetter config.TokenGetter,
	store Store) *Service {

	return &Service{
		Account: accountResolver,
		Token:   tokenGetter,
		Store:   store,
	}
}

func (s *Service) GetBalance(ctx router.Context, id *BalanceId) (*Balance, error) {
	if err := router.ValidateRequest(id); err != nil {
		return nil, err
	}

	_, err := s.Token.GetToken(ctx, &config.TokenId{Symbol: id.Symbol, Group: id.Group})
	if err != nil {
		return nil, fmt.Errorf(`get token: %w`, err)
	}
	return s.Store.Get(ctx, id)
}

func (s *Service) ListBalances(ctx router.Context, _ *emptypb.Empty) (*Balances, error) {
	// empty balance id - no conditions
	return s.ListAccountBalances(ctx, &BalanceId{})
}

func (s *Service) ListAccountBalances(ctx router.Context, id *BalanceId) (*Balances, error) {
	balances, err := s.Store.List(ctx, id)
	if err != nil {
		return nil, err
	}

	return &Balances{Items: balances}, nil
}

func (s *Service) Transfer(ctx router.Context, transfer *TransferRequest) (*TransferResponse, error) {
	if err := router.ValidateRequest(transfer); err != nil {
		return nil, err
	}

	invokerAddress, err := s.Account.GetInvokerAddress(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf(`get invoker address: %w`, err)
	}

	_, err = s.Token.GetToken(ctx, &config.TokenId{Symbol: transfer.Symbol, Group: transfer.Group})
	if err != nil {
		return nil, fmt.Errorf(`get token: %w`, err)
	}

	if err := s.Store.Transfer(ctx, &TransferOperation{
		Sender:    invokerAddress.Address,
		Recipient: transfer.Recipient,
		Symbol:    transfer.Symbol,
		Group:     transfer.Group,
		Amount:    transfer.Amount,
		Meta:      transfer.Meta,
	}); err != nil {
		return nil, err
	}

	if err = Event(ctx).Set(&Transferred{
		Sender:    invokerAddress.Address,
		Recipient: transfer.Recipient,
		Symbol:    transfer.Symbol,
		Group:     transfer.Group,
		Amount:    transfer.Amount,
	}); err != nil {
		return nil, err
	}

	return &TransferResponse{
		Sender:    invokerAddress.Address,
		Recipient: transfer.Recipient,
		Symbol:    transfer.Symbol,
		Group:     transfer.Group,
		Amount:    transfer.Amount,
	}, nil
}

func (s *Service) TransferBatch(ctx router.Context, transferBatch *TransferBatchRequest) (*TransferBatchResponse, error) {
	if err := router.ValidateRequest(transferBatch); err != nil {
		return nil, err
	}

	invokerAddress, err := s.Account.GetInvokerAddress(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf(`get invoker address: %w`, err)
	}

	// check only first
	_, err = s.Token.GetToken(ctx, &config.TokenId{
		Symbol: transferBatch.Transfers[0].Symbol, Group: transferBatch.Transfers[0].Group})
	if err != nil {
		return nil, fmt.Errorf(`get token: %w`, err)
	}

	var (
		operations []*TransferOperation
		events     []*Transferred
		responses  []*TransferResponse
	)

	for _, t := range transferBatch.Transfers {
		operations = append(operations, &TransferOperation{
			Sender:    invokerAddress.Address,
			Recipient: t.Recipient,
			Symbol:    t.Symbol,
			Group:     t.Group,
			Amount:    t.Amount,
			Meta:      t.Meta,
		})

		events = append(events, &Transferred{
			Sender:    invokerAddress.Address,
			Recipient: t.Recipient,
			Symbol:    t.Symbol,
			Group:     t.Group,
			Amount:    t.Amount,
		})

		responses = append(responses, &TransferResponse{
			Sender:    invokerAddress.Address,
			Recipient: t.Recipient,
			Symbol:    t.Symbol,
			Group:     t.Group,
			Amount:    t.Amount,
		})
	}
	if err := s.Store.TransferBatch(ctx, operations); err != nil {
		return nil, err
	}

	if err = Event(ctx).Set(&TransferredBatch{
		Transfers: events,
	}); err != nil {
		return nil, err
	}

	return &TransferBatchResponse{
		Transfers: responses,
	}, nil
}
