package token

import (
	"go-drive/common"
	"go-drive/common/registry"
	"go-drive/common/types"
)

const SessionPrefix = "s_"

func NewTokenStore(config common.Config, ch *registry.ComponentsHolder) (types.TokenStore, error) {
	switch config.TokenType {
	case FileTokenStoreName:
		return NewFileTokenStore(config, ch)
	case MemTokenStoreName:
		return NewMemTokenStore(config, ch), nil
	default:
		return NewFileTokenStore(config, ch)
	}
}