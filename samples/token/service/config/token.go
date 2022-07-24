package config

import (
	"errors"

	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/s7techlab/cckit/router"
)

var (
	ErrTokenAlreadyExists = errors.New(`token already exists`)
)

type TokenGetter interface {
	GetToken(router.Context, *TokenId) (*Token, error)
	GetDefaultToken(router.Context, *emptypb.Empty) (*Token, error)
}

func CreateDefaultToken(
	ctx router.Context, configSvc ConfigServiceChaincode, createToken *CreateTokenTypeRequest) (*TokenId, error) {

	existsTokenType, _ := configSvc.GetTokenType(ctx, &TokenTypeId{Symbol: createToken.Symbol})
	if existsTokenType != nil {
		return nil, ErrTokenAlreadyExists
	}

	// init token on first Init call
	_, err := configSvc.CreateTokenType(ctx, createToken)
	if err != nil {
		return nil, err
	}

	tokenId := &TokenId{
		Symbol: createToken.Symbol,
		Group:  nil,
	}

	if _, err = configSvc.SetConfig(ctx, &Config{DefaultToken: tokenId}); err != nil {
		return nil, err
	}

	return tokenId, nil
}
