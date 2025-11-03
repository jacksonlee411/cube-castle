package authbff

import (
    "sync"
    "time"
)

// Session 服务器侧会话（仅保存刷新令牌/用户与租户信息）
type Session struct {
    ID         string
    UserID     string
    UserName   string
    UserEmail  string
    TenantID   string
    Roles      []string
    Scopes     []string
    RefreshTok string
    IDToken    string
    CreatedAt  time.Time
    LastUsedAt time.Time
    ExpiresAt  time.Time // 会话总体过期（如30天）
}

// Store 会话存储接口（默认内存实现，后续可切换 Redis）
type Store interface {
    Set(*Session)
    Get(id string) (*Session, bool)
    Delete(id string)
}

// InMemoryStore 简单内存实现（单实例）
type InMemoryStore struct {
    mu   sync.RWMutex
    data map[string]*Session
}

func NewInMemoryStore() *InMemoryStore {
    return &InMemoryStore{data: make(map[string]*Session)}
}

func (s *InMemoryStore) Set(sess *Session) {
    s.mu.Lock(); defer s.mu.Unlock()
    s.data[sess.ID] = sess
}

func (s *InMemoryStore) Get(id string) (*Session, bool) {
    s.mu.RLock(); defer s.mu.RUnlock()
    v, ok := s.data[id]
    return v, ok
}

func (s *InMemoryStore) Delete(id string) {
    s.mu.Lock(); defer s.mu.Unlock()
    delete(s.data, id)
}
