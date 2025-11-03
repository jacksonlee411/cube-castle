package authbff

import (
    "sync"
    "time"
)

// AuthFlowState 临时保存授权流程材料（PKCE/state/nonce/redirect）
type AuthFlowState struct {
    State        string
    Nonce        string
    CodeVerifier string
    RedirectPath string
    CreatedAt    time.Time
    ExpiresAt    time.Time
}

// AuthFlowStore 简易内存存储（按 state 管理，带TTL）
type AuthFlowStore struct {
    mu   sync.RWMutex
    data map[string]*AuthFlowState
}

func NewAuthFlowStore() *AuthFlowStore {
    return &AuthFlowStore{data: map[string]*AuthFlowState{}}
}

func (s *AuthFlowStore) Set(f *AuthFlowState) {
    s.mu.Lock(); defer s.mu.Unlock()
    s.data[f.State] = f
}

func (s *AuthFlowStore) Get(state string) (*AuthFlowState, bool) {
    s.mu.RLock(); defer s.mu.RUnlock()
    f, ok := s.data[state]
    if !ok { return nil, false }
    if time.Now().After(f.ExpiresAt) { return nil, false }
    return f, true
}

func (s *AuthFlowStore) Delete(state string) {
    s.mu.Lock(); defer s.mu.Unlock()
    delete(s.data, state)
}

