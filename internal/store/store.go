package store

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Store struct {
	mu     sync.RWMutex
	users  map[string]*UserData // userID -> data
	tokens map[string]string    // token -> userID
}

type UserData struct {
	User   User
	Traces map[string]*StoredTrace
}

type StoredTrace struct {
	ID        string
	RawJSON   json.RawMessage
	CreatedAt time.Time
}

type User struct {
	ID    string
	Token string
}

func NewStore() *Store {
	return &Store{
		users:  make(map[string]*UserData),
		tokens: make(map[string]string),
	}
}

func (s *Store) GetOrCreateUser(userID string) *UserData {
	s.mu.Lock()
	defer s.mu.Unlock()

	userData, ok := s.users[userID]
	if !ok {
		NewUser := s.RegisterUser()
		userData = &UserData{
			User:   NewUser,
			Traces: make(map[string]*StoredTrace),
		}
		s.users[NewUser.ID] = userData
	}
	return userData
}

func (s *Store) AddTrace(userID string, trace *StoredTrace) {
	user := s.GetOrCreateUser(userID)

	s.mu.Lock()
	defer s.mu.Unlock()

	user.Traces[trace.ID] = trace
}

func (s *Store) GetTrace(userID, traceID string) (*StoredTrace, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, ok := s.users[userID]
	if !ok {
		return nil, false
	}

	trace, ok := user.Traces[traceID]
	return trace, ok
}

func (s *Store) RegisterUser() User {
	s.mu.Lock()
	defer s.mu.Unlock()

	userID := uuid.NewString()
	token := uuid.NewString()

	user := User{
		ID:    userID,
		Token: token,
	}

	s.users[userID] = &UserData{
		User:   user,
		Traces: make(map[string]*StoredTrace),
	}

	s.tokens[token] = userID

	return user
}

func (s *Store) UserFromToken(token string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	userID, ok := s.tokens[token]
	return userID, ok
}
