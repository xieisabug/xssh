package forwarding

import (
	"fmt"
	"os"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
	"xssh/internal/config"
)

// ForwardingManager manages all port forwarding sessions
type ForwardingManager struct {
	sessions sync.Map // map[string]*ForwardingSession
	sshClients sync.Map // map[string]*ssh.Client for SSH connections
	mu       sync.RWMutex
}

// NewManager creates a new forwarding manager
func NewManager() *ForwardingManager {
	return &ForwardingManager{}
}

// StartForwarding starts a new port forwarding session
func (fm *ForwardingManager) StartForwarding(rule ForwardingRule, host config.SSHHost, keyPassword string) error {
	// Check if session already exists
	if _, exists := fm.sessions.Load(rule.ID); exists {
		return fmt.Errorf("forwarding session %s already exists", rule.ID)
	}

	// Create new session
	session := &ForwardingSession{
		Rule: rule,
		Stats: ForwardingStats{
			StartTime: time.Now(),
		},
		done: make(chan struct{}),
	}

	// Store session
	fm.sessions.Store(rule.ID, session)

	// Start the appropriate forwarding type
	var err error
	switch rule.Type {
	case LocalForward:
		err = fm.startLocalForwarding(session, host, keyPassword)
	case RemoteForward:
		err = fm.startRemoteForwarding(session, host, keyPassword)
	case DynamicForward:
		err = fm.startDynamicForwarding(session, host, keyPassword)
	default:
		err = fmt.Errorf("unsupported forwarding type: %v", rule.Type)
	}

	if err != nil {
		fm.sessions.Delete(rule.ID)
		return err
	}

	session.SetActive(true)
	return nil
}

// StopForwarding stops a port forwarding session
func (fm *ForwardingManager) StopForwarding(sessionID string) error {
	sessionInterface, exists := fm.sessions.Load(sessionID)
	if !exists {
		return fmt.Errorf("session %s not found", sessionID)
	}

	session := sessionInterface.(*ForwardingSession)
	session.SetActive(false)

	// Close listener if exists
	if session.listener != nil {
		session.listener.Close()
	}

	// Signal shutdown
	close(session.done)

	// Remove from sessions
	fm.sessions.Delete(sessionID)

	return nil
}

// GetSession retrieves a forwarding session by ID
func (fm *ForwardingManager) GetSession(sessionID string) (*ForwardingSession, bool) {
	sessionInterface, exists := fm.sessions.Load(sessionID)
	if !exists {
		return nil, false
	}
	return sessionInterface.(*ForwardingSession), true
}

// GetAllSessions returns all active forwarding sessions
func (fm *ForwardingManager) GetAllSessions() []*ForwardingSession {
	var sessions []*ForwardingSession
	fm.sessions.Range(func(key, value interface{}) bool {
		session := value.(*ForwardingSession)
		sessions = append(sessions, session)
		return true
	})
	return sessions
}

// StopAll stops all forwarding sessions
func (fm *ForwardingManager) StopAll() {
	var sessionIDs []string
	fm.sessions.Range(func(key, value interface{}) bool {
		sessionIDs = append(sessionIDs, key.(string))
		return true
	})

	for _, id := range sessionIDs {
		fm.StopForwarding(id)
	}
}

// GetSSHClient gets or creates an SSH client for the host
func (fm *ForwardingManager) getSSHClient(host config.SSHHost, keyPassword string) (*ssh.Client, error) {
	clientKey := fmt.Sprintf("%s@%s:%s", host.User, host.Host, host.Port)
	
	// Check if client already exists
	if clientInterface, exists := fm.sshClients.Load(clientKey); exists {
		client := clientInterface.(*ssh.Client)
		// Test if connection is still alive
		_, _, err := client.SendRequest("keepalive@golang.org", true, nil)
		if err == nil {
			return client, nil
		}
		// Connection is dead, remove it
		fm.sshClients.Delete(clientKey)
		client.Close()
	}

	// Create new SSH client
	client, err := fm.createSSHClient(host, keyPassword)
	if err != nil {
		return nil, err
	}

	fm.sshClients.Store(clientKey, client)
	return client, nil
}

// createSSHClient creates a new SSH client connection
func (fm *ForwardingManager) createSSHClient(host config.SSHHost, keyPassword string) (*ssh.Client, error) {
	var auth []ssh.AuthMethod

	if host.Identity != "" {
		// Use key-based authentication
		key, err := fm.loadPrivateKey(host.Identity, keyPassword)
		if err != nil {
			return nil, fmt.Errorf("failed to load private key: %v", err)
		}
		auth = append(auth, ssh.PublicKeys(key))
	}

	config := &ssh.ClientConfig{
		User:            host.User,
		Auth:            auth,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}

	client, err := ssh.Dial("tcp", host.Host+":"+host.Port, config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to SSH server: %v", err)
	}

	return client, nil
}

// loadPrivateKey loads and parses a private key with optional password
func (fm *ForwardingManager) loadPrivateKey(keyPath, keyPassword string) (ssh.Signer, error) {
	keyData, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, err
	}

	var key ssh.Signer
	if keyPassword != "" {
		key, err = ssh.ParsePrivateKeyWithPassphrase(keyData, []byte(keyPassword))
	} else {
		key, err = ssh.ParsePrivateKey(keyData)
	}

	return key, err
}