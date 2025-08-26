package sessions

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"net/http"
	"sync"
	"time"
)

const (
	cookieName = "sid"
	cookieTTL  = 30 * 24 * time.Hour
)

type Store struct {
	Mu         *sync.RWMutex
	SessionMap map[string]*Session
}

func NewStore() *Store {
	var mutex sync.RWMutex
	return &Store{SessionMap: make(map[string]*Session), Mu: &mutex}
}

func (s *Store) EnsureSessionID(w http.ResponseWriter, r *http.Request) string {
	cookie, err := r.Cookie(cookieName)
	if err == nil && cookie.Value != "" {
		s.Mu.Lock()
		if _, exists := s.SessionMap[cookie.Value]; !exists {
			s.SessionMap[cookie.Value] = &Session{ID: cookie.Value, LastSeen: time.Now()}
		}
		s.Mu.Unlock()
		return cookie.Value
	}
	sid := randomB64(32)
	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    sid,
		Path:     "/",
		Expires:  time.Now().Add(cookieTTL),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		//Secure: true(in prod),
	})
	s.Mu.Lock()
	if _, exists := s.SessionMap[sid]; !exists {
		s.SessionMap[sid] = &Session{ID: sid, LastSeen: time.Now()}
	}
	s.Mu.Unlock()
	return sid
}

func (s *Store) GetSessionSnapshot(sid string) (Session, bool) {
	s.Mu.RLock()
	defer s.Mu.RUnlock()
	session, exists := s.SessionMap[sid]
	if !exists {
		return Session{}, false
	}
	return *session, exists
}

func (s *Store) SetSessionToken(sid string, token Tokens) error {
	s.Mu.Lock()
	defer s.Mu.Unlock()
	_, exists := s.SessionMap[sid]
	if !exists {
		return errors.New("no existing session for session token")
	}
	s.SessionMap[sid].Tokens = &token
	return nil
}

func (s *Store) Save(sid string, session *Session) {
	session.LastSeen = time.Now()
	s.Mu.Lock()
	s.SessionMap[sid] = session
	s.Mu.Unlock()
}

func (s *Store) Delete(sid string) {
	s.Mu.Lock()
	defer s.Mu.Unlock()
	delete(s.SessionMap, sid)
}

func randomB64(length int) string {
	byteSlice := make([]byte, length)
	_, err := rand.Read(byteSlice)
	if err != nil {
		panic(err)
	}
	return base64.RawURLEncoding.EncodeToString(byteSlice)
}

func (s *Store) SetPendingAuth(sid string, pendingAuth PendingAuth) error {
	s.Mu.Lock()
	defer s.Mu.Unlock()
	_, exists := s.SessionMap[sid]
	if !exists {
		return errors.New("no existing session for pending auth")
	}
	s.SessionMap[sid].PendingAuth = &pendingAuth
	return nil
}

func (s *Store) RemovePendingAuth(sid string) error {
	s.Mu.Lock()
	defer s.Mu.Unlock()
	_, exists := s.SessionMap[sid]
	if !exists {
		return errors.New("no existing session for pending auth")
	}
	s.SessionMap[sid].PendingAuth = nil
	return nil
}

func (s *Store) GetPendingAuthCopy(sid string) (PendingAuth, bool, error) {
	s.Mu.RLock()
	defer s.Mu.RUnlock()
	session, exists := s.SessionMap[sid]
	if !exists {
		return PendingAuth{}, false, errors.New("no existing session for pending auth")
	}
	if session.PendingAuth == nil {
		return PendingAuth{}, false, errors.New("no pending auth exists")
	}
	return *session.PendingAuth, true, nil
}
func (s *Store) GetTokensCopy(sid string) (Tokens, bool, error) {
	s.Mu.RLock()
	defer s.Mu.RUnlock()
	session, exists := s.SessionMap[sid]
	if !exists {
		return Tokens{}, false, errors.New("no existing session for token")
	}
	if session.Tokens == nil {
		return Tokens{}, false, errors.New("no token exists")
	}
	return *session.Tokens, true, nil
}

func (s *Store) SetTokenExpiration(sid string, time time.Time) bool {
	s.Mu.Lock()
	defer s.Mu.Unlock()
	session, exists := s.SessionMap[sid]
	if !exists {
		return false
	}
	session.Tokens.ExpiresAt = time
	return true
}

// func (s *Store) GetUserCopy(sid string) (UserProfile, bool, error)

type Session struct {
	ID          string
	PendingAuth *PendingAuth
	Tokens      *Tokens
	User        *UserProfile
	LastSeen    time.Time
}

type PendingAuth struct {
	State        string
	CodeVerifier string
	Scopes       []string
	CreatedAt    time.Time
	ExpiresAt    time.Time
}

type Tokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	Scope        string `json:"scope"`
	TokenType    string `json:"token_type"`
	ExpiresAt    time.Time
}

type UserProfile struct {
	ID          string
	DisplayName string
	Email       string
}
