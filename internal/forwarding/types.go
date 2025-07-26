package forwarding

import (
	"net"
	"sync/atomic"
	"time"
)

// ForwardingType represents the type of port forwarding
type ForwardingType int

const (
	LocalForward ForwardingType = iota  // -L: Local port to remote host:port
	RemoteForward                       // -R: Remote port to local host:port
	DynamicForward                      // -D: SOCKS5 proxy
)

func (ft ForwardingType) String() string {
	switch ft {
	case LocalForward:
		return "Local"
	case RemoteForward:
		return "Remote"
	case DynamicForward:
		return "Dynamic"
	default:
		return "Unknown"
	}
}

// ForwardingRule represents a port forwarding configuration
type ForwardingRule struct {
	ID          string         // Unique identifier
	Type        ForwardingType // Type of forwarding
	LocalHost   string         // Local host (usually "localhost" or "0.0.0.0")
	LocalPort   int            // Local port
	RemoteHost  string         // Remote host
	RemotePort  int            // Remote port
	Description string         // User description
}

// ForwardingStats holds statistics for a forwarding session
type ForwardingStats struct {
	BytesReceived    int64     // Total bytes received
	BytesSent        int64     // Total bytes sent
	ConnectionCount  int64     // Number of connections handled
	ActiveConnections int64    // Current active connections
	StartTime        time.Time // When the forwarding started
	LastActivity     time.Time // Last data transfer time
	ErrorCount       int64     // Number of errors encountered
	LastError        string    // Last error message
}

// ForwardingSession represents an active port forwarding session
type ForwardingSession struct {
	Rule     ForwardingRule // The forwarding rule
	Stats    ForwardingStats // Statistics
	listener net.Listener   // The listener for the session
	done     chan struct{}  // Channel to signal shutdown
	active   int32          // Atomic flag for active state
}

// IsActive returns whether the session is currently active
func (fs *ForwardingSession) IsActive() bool {
	return atomic.LoadInt32(&fs.active) == 1
}

// SetActive sets the active state of the session
func (fs *ForwardingSession) SetActive(active bool) {
	if active {
		atomic.StoreInt32(&fs.active, 1)
	} else {
		atomic.StoreInt32(&fs.active, 0)
	}
}

// AddBytesReceived atomically adds to bytes received
func (fs *ForwardingSession) AddBytesReceived(bytes int64) {
	atomic.AddInt64(&fs.Stats.BytesReceived, bytes)
	fs.Stats.LastActivity = time.Now()
}

// AddBytesSent atomically adds to bytes sent
func (fs *ForwardingSession) AddBytesSent(bytes int64) {
	atomic.AddInt64(&fs.Stats.BytesSent, bytes)
	fs.Stats.LastActivity = time.Now()
}

// IncrementConnections atomically increments connection count
func (fs *ForwardingSession) IncrementConnections() {
	atomic.AddInt64(&fs.Stats.ConnectionCount, 1)
	atomic.AddInt64(&fs.Stats.ActiveConnections, 1)
}

// DecrementActiveConnections atomically decrements active connection count
func (fs *ForwardingSession) DecrementActiveConnections() {
	atomic.AddInt64(&fs.Stats.ActiveConnections, -1)
}

// IncrementErrors atomically increments error count
func (fs *ForwardingSession) IncrementErrors(err string) {
	atomic.AddInt64(&fs.Stats.ErrorCount, 1)
	fs.Stats.LastError = err
}

// GetUptime returns the duration since the session started
func (fs *ForwardingSession) GetUptime() time.Duration {
	return time.Since(fs.Stats.StartTime)
}

// GetTransferRate returns the current transfer rate in bytes per second
func (fs *ForwardingSession) GetTransferRate() (float64, float64) {
	uptime := fs.GetUptime().Seconds()
	if uptime == 0 {
		return 0, 0
	}
	
	received := float64(atomic.LoadInt64(&fs.Stats.BytesReceived))
	sent := float64(atomic.LoadInt64(&fs.Stats.BytesSent))
	
	return received / uptime, sent / uptime
}