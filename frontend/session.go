package frontend

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type SessionManager struct {
	lock        sync.Mutex // protects session
	lastId      int64
	maxLifeTime int
	sessions    map[int64]*Session
}

func NewSessionManager(maxlifetime int) *SessionManager {
	manager := &SessionManager{maxLifeTime: maxlifetime}
	manager.sessions = make(map[int64]*Session)
	return manager
}

func (m *SessionManager) StartSession(nonceSize int) *Session {
	m.lock.Lock()
	defer m.lock.Unlock()

	session := m.initNewSession(nonceSize)
	m.sessions[session.GetID()] = session
	return session
}

func (m SessionManager) GetSession(sessionID int64) *Session {
	session, _ := m.sessions[sessionID]
	return session
}

func (m *SessionManager) EndSession(sessionId int64) error {
	if _, ok := m.sessions[sessionId]; !ok {
		fmt.Errorf("session with id \"%d\" does not exist", sessionId)
	}

	delete(m.sessions, sessionId)
	return nil
}

func (m *SessionManager) generateSessionID() int64 {
	m.lastId++
	return m.lastId
}

func (m *SessionManager) initNewSession(nonceSize int) *Session {
	id := m.generateSessionID()
	s := new(Session)

	nonce := make([]byte, nonceSize)
	rand.Read(nonce)

	expiry := time.Now().Add(time.Second * time.Duration(m.maxLifeTime))

	s.Init(id, nonce, expiry)
	return s
}

type Session struct {
	SessionInfo
	id   int64
	data map[interface{}]interface{}
}

type SessionInfo struct {
	Nonce  []byte    `json:"nonce", binding:"required"`
	Expiry time.Time `json:"expiry", binding:"required"`
	Accept []string  `json:"accept", binding:"required"`
	State  string    `json:"state", binding:"required"`
}

// Init initializes a new session with the specified ID
func (s *Session) Init(id int64, nonce []byte, expiry time.Time) {
	s.Nonce = nonce
	s.Expiry = expiry
	s.Accept = []string{ // TODO: should not be hard-coded
		"application/psa-attestation-token",
	}
	s.State = "waiting"
	s.id = id
	s.data = make(map[interface{}]interface{})
}

// GetID returns the ID of this session
func (s Session) GetID() int64 {
	return s.id
}

// Set sets the specified key to the specified value within the session
func (s Session) Set(key, value interface{}) error {
	if _, ok := s.data[key]; ok {
		return fmt.Errorf("key \"%v\" already exists for session %d", key, s.id)
	}

	s.data[key] = value
	return nil
}

// Get returns the value for the specified key, or nil if the keys is not set in this session.
func (s Session) Get(key interface{}) interface{} {
	value, _ := s.data[key]
	return value
}

// Delete removes the specified key and the associated value from the session's data.
func (s *Session) Delete(key interface{}) error {
	if _, ok := s.data[key]; !ok {
		return fmt.Errorf("cannot delete key \"%v\" in session %d - does not exist", key, s.id)
	}

	delete(s.data, key)
	return nil
}

func (s *Session) SetState(value string) {
	s.SessionInfo.State = value
}
