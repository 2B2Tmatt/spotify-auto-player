package sessions

import (
	"crypto/rand"
	"encoding/base64"
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

func (s *Store) Get(sid string) (*Session, bool) {
	s.Mu.RLock()
	defer s.Mu.RUnlock()
	session, exists := s.SessionMap[sid]
	return session, exists
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

func (s *Store) AddPendingAuth(sid string, pending *PendingAuth) {
	s.Mu.Lock()
	defer s.Mu.Unlock()
	s.SessionMap[sid].PendingAuth = pending
}

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
	AccessToken  string        `json:"access_token"`
	RefreshToken string        `json:"refresh_token"`
	ExpiresIn    time.Duration `json:"expires_in"`
	Scope        string        `json:"scope"`
	TokenType    string        `json:"token_type"`
}

type UserProfile struct {
	ID          string
	DisplayName string
	Email       string
}
