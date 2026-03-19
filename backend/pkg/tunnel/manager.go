package tunnel

import (
	"errors"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/hashicorp/yamux"
)

var (
	ErrWorkerOffline = errors.New("worker tunnel is offline")
	
	// GlobalManager holds the singleton instance of the tunnel manager since it needs
	// to span across the User API and Worker Management Router contexts
	GlobalManager *Manager
)

func init() {
	GlobalManager = NewManager()
}

// Manager stores all active multiplexed Yamux sessions connected from Worker nodes.
type Manager struct {
	mu       sync.RWMutex
	sessions map[string]*yamux.Session // Maps WorkerID to their Yamux Session
}

// NewManager creates a new tunnel manager.
func NewManager() *Manager {
	return &Manager{
		sessions: make(map[string]*yamux.Session),
	}
}

// Register upgrades a websocket to a yamux session and registers the worker.
// The session acts as a server on the Yamux layer, accepting reverse streams 
// from the worker, or opening streams directly to the worker.
func (m *Manager) Register(workerID string, ws *websocket.Conn) (*yamux.Session, error) {
	netConn := NewWSConn(ws)
	
	// We are the server, the Worker node is the client dialing in to us.
	// But in Yamux, either side can open a stream once established.
	config := yamux.DefaultConfig()
	// Enable keepalive to prevent firewall drops
	config.EnableKeepAlive = true
	
	session, err := yamux.Server(netConn, config)
	if err != nil {
		netConn.Close()
		return nil, err
	}

	m.mu.Lock()
	// Terminate previous if exists (handling unclean disconnects)
	if existing, ok := m.sessions[workerID]; ok {
		existing.Close()
	}
	m.sessions[workerID] = session
	m.mu.Unlock()

	// Wait for the session to close on its own and auto-deregister
	go func() {
		<-session.CloseChan()
		m.Unregister(workerID, session)
	}()

	return session, nil
}

// Unregister safely removes a session if it matches the provided instance.
func (m *Manager) Unregister(workerID string, session *yamux.Session) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if current, ok := m.sessions[workerID]; ok && current == session {
		delete(m.sessions, workerID)
	}
}

// GetSession retrieves the active Yamux session for a worker to open a proxied transport.
func (m *Manager) GetSession(workerID string) (*yamux.Session, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	session, ok := m.sessions[workerID]
	if !ok || session.IsClosed() {
		return nil, ErrWorkerOffline
	}

	return session, nil
}
