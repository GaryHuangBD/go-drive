package token

import (
	"fmt"
	"github.com/google/uuid"
	cmap "github.com/orcaman/concurrent-map"
	"go-drive/common"
	"go-drive/common/errors"
	"go-drive/common/i18n"
	"go-drive/common/registry"
	"go-drive/common/types"
	"go-drive/common/utils"
	"log"
	"sync"
	"time"
)

const MemTokenStoreName = "mem"

type MemTokenStore struct {
	store cmap.ConcurrentMap

	validity    time.Duration
	autoRefresh bool

	mux *sync.Mutex

	tickerStop func()
}

// NewMemTokenStore creates a MemTokenStore
//
// params:
//
// - autoRefresh: refresh token by adding `validity` after each token access
//
// - cleanupDuration: cleanup invalid token each `cleanupDuration`
func NewMemTokenStore(config common.Config, ch *registry.ComponentsHolder) types.TokenStore {

	tokenStore := &MemTokenStore{
		store:       cmap.New(),
		validity:    config.TokenValidity,
		autoRefresh: config.TokenRefresh,
		mux:         &sync.Mutex{},
	}

	tokenStore.tickerStop = utils.TimeTick(tokenStore.clean, config.TokenValidity)
	ch.Add("tokenStore", tokenStore)

	return tokenStore
}

func (m *MemTokenStore) Create(value types.Session) (types.Token, error) {
	key := uuid.New().String()
	var expiredAt int64 = -1
	if m.validity > 0 {
		expiredAt = time.Now().Add(m.validity).Unix()
	}
	token := types.Token{Token: key, Value: value, ExpiredAt: expiredAt}
	m.store.Set(key, token)
	return token, nil
}

func (m *MemTokenStore) Update(token string, value types.Session) (types.Token, error) {
	m.mux.Lock()
	defer m.mux.Unlock()
	t, ok := m.store.Get(token)
	if !ok {
		return types.Token{}, err.NewUnauthorizedError(i18n.T("api.mem_token.invalid_token"))
	}
	tt := t.(types.Token)
	tt.Value = value
	if m.refreshEnabled() {
		tt.ExpiredAt = time.Now().Add(m.validity).Unix()
	}
	m.store.Set(token, tt)
	return tt, nil
}

func (m *MemTokenStore) Validate(token string) (types.Token, error) {
	if m.refreshEnabled() {
		m.mux.Lock()
		defer m.mux.Unlock()
	}
	t, ok := m.store.Get(token)
	if !ok {
		return types.Token{}, err.NewUnauthorizedError(i18n.T("api.mem_token.invalid_token"))
	}
	tt := t.(types.Token)
	if !m.isValid(tt) {
		return types.Token{}, err.NewUnauthorizedError(i18n.T("api.mem_token.invalid_token"))
	}
	if m.refreshEnabled() {
		tt.ExpiredAt = time.Now().Add(m.validity).Unix()
		m.store.Set(token, tt)
	}
	return tt, nil
}

func (m *MemTokenStore) Revoke(token string) error {
	m.store.Remove(token)
	return nil
}

func (m *MemTokenStore) isValid(token types.Token) bool {
	return token.ExpiredAt <= 0 || token.ExpiredAt > time.Now().Unix()
}

func (m *MemTokenStore) refreshEnabled() bool {
	return m.autoRefresh && m.validity > 0
}

func (m *MemTokenStore) clean() {
	keys := make([]string, 0)
	m.store.IterCb(func(key string, v interface{}) {
		if !m.isValid(v.(types.Token)) {
			keys = append(keys, key)
		}
	})
	for _, key := range keys {
		_ = m.Revoke(key)
	}
	log.Printf("%d expired tokens cleaned", len(keys))
}

func (m *MemTokenStore) Dispose() error {
	m.tickerStop()
	return nil
}

func (m *MemTokenStore) Status() (string, types.SM, error) {
	total := 0
	active := 0
	keys := m.store.Keys()
	for i:= 0; i < len(keys); i++ {
		total++
		t, ok := m.store.Get(keys[i])
		if ok {
			tt := t.(types.Token)
			if time.Now().Unix() < tt.ExpiredAt {
				active++
			}
		}
	}

	return "Session", types.SM{
		"Total":  fmt.Sprintf("%d", total),
		"Active": fmt.Sprintf("%d", active),
	}, nil
}
