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

func (s *Service) Transfer(ctx router.Context, req *TransferRequest) (*TransferResponse, error) {
	if err := router.ValidateRequest(req); err != nil {
		return nil, err
	}

	invokerAddress, err := s.Account.GetInvokerAddress(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf(`get invoker address: %w`, err)
	}

	_, err = s.Token.GetToken(ctx, &config.TokenId{Symbol: req.Symbol, Group: req.Group})
	if err != nil {
		return nil, fmt.Errorf(`get token: %w`, err)
	}

	if err := s.Store.Transfer(ctx, &TransferOperation{
		Sender:    invokerAddress.Address,
		Recipient: req.Recipient,
		Symbol:    req.Symbol,
		Group:     req.Group,
		Amount:    req.Amount,
		Meta:      nil,
	}); err != nil {
		return nil, err
	}

	if err = Event(ctx).Set(&Transferred{
		Sender:    invokerAddress.Address,
		Recipient: req.Recipient,
		Symbol:    req.Symbol,
		Group:     req.Group,
		Amount:    req.Amount,
	}); err != nil {
		return nil, err
	}

	return &TransferResponse{
		Sender:    invokerAddress.Address,
		Recipient: req.Recipient,
		Symbol:    req.Symbol,
		Group:     req.Group,
		Amount:    req.Amount,
	}, nil
}
