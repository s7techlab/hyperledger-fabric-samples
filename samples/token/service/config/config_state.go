package config

import (
	"fmt"

	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/s7techlab/cckit/router"
)

type StateService struct {
}

func NewStateService() *StateService {
	return &StateService{}
}

func (s *StateService) SetConfig(ctx router.Context, config *Config) (*Config, error) {
	err := State(ctx).Put(config)
	if err != nil {
		return nil, err
	}
	return config, nil
}

func (s *StateService) GetConfig(ctx router.Context, _ *emptypb.Empty) (*Config, error) {
	config, err := State(ctx).Get(&Config{}, &Config{})
	if err != nil {
		return nil, err
	}
	return config.(*Config), nil
}

// GetToken naming for token is[{TokenType}, {GroupIdPart1}, {GroupIdPart2}]
func (s *StateService) GetToken(ctx router.Context, id *TokenId) (*Token, error) {
	if err := router.ValidateRequest(id); err != nil {
		return nil, err
	}
	var (
		tokenType  *TokenType
		tokenGroup *TokenGroup
		err        error
	)

	tokenType, err = s.GetTokenType(ctx, &TokenTypeId{Symbol: id.Symbol})
	if err != nil {
		return nil, fmt.Errorf(`token type: %w`, err)
	}

	if len(id.Group) > 0 {
		tokenGroup, err = s.GetTokenGroup(ctx, &TokenGroupId{Symbol: id.Symbol, Group: id.Group})
		if err != nil {
			return nil, fmt.Errorf(`token type: %w`, err)
		}
	}

	return &Token{
		Type:  tokenType,
		Group: tokenGroup,
	}, nil
}

func (s *StateService) GetDefaultToken(ctx router.Context, _ *emptypb.Empty) (*Token, error) {
	config, err := s.GetConfig(ctx, nil)
	if err != nil {
		return nil, err
	}

	if config.DefaultToken != nil {
		return s.GetToken(ctx, config.DefaultToken)
	}

	return nil, nil
}

func (s *StateService) CreateTokenType(ctx router.Context, req *CreateTokenTypeRequest) (*TokenType, error) {
	if err := router.ValidateRequest(req); err != nil {
		return nil, err
	}

	tokenType := &TokenType{
		Name:        req.Name,
		Symbol:      req.Symbol,
		Decimals:    req.Decimals,
		TotalSupply: req.TotalSupply,
		GroupType:   req.GroupType,
	}

	for _, m := range req.Meta {
		tokenType.Meta = append(tokenType.Meta, &TokenMeta{
			Key:   m.Key,
			Value: m.Value,
		})
	}
	if err := State(ctx).Insert(tokenType); err != nil {
		return nil, err
	}

	if err := Event(ctx).Set(&TokenTypeCreated{
		Symbol: req.Symbol,
		Name:   req.Name,
	}); err != nil {
		return nil, err
	}
	return tokenType, nil
}

func (s *StateService) GetTokenType(ctx router.Context, id *TokenTypeId) (*TokenType, error) {
	if err := router.ValidateRequest(id); err != nil {
		return nil, err
	}

	tokenType, err := State(ctx).Get(id, &TokenType{})
	if err != nil {
		return nil, err
	}
	return tokenType.(*TokenType), nil
}

func (s *StateService) ListTokenTypes(ctx router.Context, _ *emptypb.Empty) (*TokenTypes, error) {
	tokenTypes, err := State(ctx).List(&TokenType{})
	if err != nil {
		return nil, err
	}
	return tokenTypes.(*TokenTypes), nil
}

func (s *StateService) UpdateTokenType(ctx router.Context, request *UpdateTokenTypeRequest) (*TokenType, error) {
	//TODO implement me
	panic("implement me")
}

func (s *StateService) DeleteTokenType(ctx router.Context, id *TokenTypeId) (*TokenType, error) {
	//TODO implement me
	panic("implement me")
}

func (s *StateService) GetTokenGroups(ctx router.Context, id *TokenTypeId) (*TokenGroups, error) {
	if err := router.ValidateRequest(id); err != nil {
		return nil, err
	}

	tokenGroups, err := State(ctx).ListWith(&TokenGroup{}, []string{id.Symbol})
	if err != nil {
		return nil, err
	}
	return tokenGroups.(*TokenGroups), nil
}

func (s *StateService) CreateTokenGroup(ctx router.Context, req *CreateTokenGroupRequest) (*TokenGroup, error) {
	if err := router.ValidateRequest(req); err != nil {
		return nil, err
	}

	_, err := s.GetTokenType(ctx, &TokenTypeId{Symbol: req.Symbol})
	if err != nil {
		return nil, err
	}

	tokenGroup := &TokenGroup{
		Symbol:      req.Symbol,
		Group:       req.Group,
		Name:        req.Name,
		TotalSupply: 0,
	}

	for _, m := range req.Meta {
		tokenGroup.Meta = append(tokenGroup.Meta, &TokenMeta{
			Key:   m.Key,
			Value: m.Value,
		})
	}
	if err := State(ctx).Insert(tokenGroup); err != nil {
		return nil, err
	}

	if err := Event(ctx).Set(&TokenGroupCreated{
		Symbol: req.Symbol,
		Group:  req.Group,
		Name:   req.Name,
	}); err != nil {
		return nil, err
	}
	return tokenGroup, nil
}

func (s *StateService) GetTokenGroup(ctx router.Context, id *TokenGroupId) (*TokenGroup, error) {
	tokenGroup, err := State(ctx).Get(id, &TokenGroup{})
	if err != nil {
		return nil, err
	}
	return tokenGroup.(*TokenGroup), nil
}

func (s *StateService) DeleteTokenGroup(ctx router.Context, id *TokenGroupId) (*Token, error) {
	//TODO implement me
	panic("implement me")
}
