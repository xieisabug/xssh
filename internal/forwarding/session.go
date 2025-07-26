package forwarding

import (
	"fmt"
	"io"
	"net"
	"time"

	"golang.org/x/crypto/ssh"
	"xssh/internal/config"
)

// startLocalForwarding implements local port forwarding (-L)
// Listens on local port and forwards connections to remote host:port through SSH
func (fm *ForwardingManager) startLocalForwarding(session *ForwardingSession, host config.SSHHost, keyPassword string) error {
	rule := session.Rule
	
	// Get SSH client
	sshClient, err := fm.getSSHClient(host, keyPassword)
	if err != nil {
		return fmt.Errorf("failed to get SSH client: %v", err)
	}

	// Listen on local port
	localAddr := fmt.Sprintf("%s:%d", rule.LocalHost, rule.LocalPort)
	listener, err := net.Listen("tcp", localAddr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %v", localAddr, err)
	}

	session.listener = listener

	// Start accepting connections in a goroutine
	go func() {
		defer listener.Close()
		
		for {
			select {
			case <-session.done:
				return
			default:
				// Accept connection with timeout
				if tcpListener, ok := listener.(*net.TCPListener); ok {
					tcpListener.SetDeadline(time.Now().Add(1 * time.Second))
				}
				
				localConn, err := listener.Accept()
				if err != nil {
					if ne, ok := err.(net.Error); ok && ne.Timeout() {
						continue // Timeout is expected for graceful shutdown
					}
					if session.IsActive() {
						session.IncrementErrors(fmt.Sprintf("Accept error: %v", err))
					}
					continue
				}

				// Handle connection in separate goroutine
				go fm.handleLocalForwardConnection(session, sshClient, localConn, rule.RemoteHost, rule.RemotePort)
			}
		}
	}()

	return nil
}

// handleLocalForwardConnection handles a single local forward connection
func (fm *ForwardingManager) handleLocalForwardConnection(session *ForwardingSession, sshClient *ssh.Client, localConn net.Conn, remoteHost string, remotePort int) {
	defer localConn.Close()
	
	session.IncrementConnections()
	defer session.DecrementActiveConnections()

	// Connect to remote host through SSH
	remoteAddr := fmt.Sprintf("%s:%d", remoteHost, remotePort)
	remoteConn, err := sshClient.Dial("tcp", remoteAddr)
	if err != nil {
		session.IncrementErrors(fmt.Sprintf("Failed to connect to %s: %v", remoteAddr, err))
		return
	}
	defer remoteConn.Close()

	// Start data forwarding
	fm.forwardData(session, localConn, remoteConn)
}

// startRemoteForwarding implements remote port forwarding (-R)
// Listens on remote port and forwards connections to local host:port
func (fm *ForwardingManager) startRemoteForwarding(session *ForwardingSession, host config.SSHHost, keyPassword string) error {
	rule := session.Rule
	
	// Get SSH client
	sshClient, err := fm.getSSHClient(host, keyPassword)
	if err != nil {
		return fmt.Errorf("failed to get SSH client: %v", err)
	}

	// Listen on remote port through SSH
	remoteAddr := fmt.Sprintf("%s:%d", rule.RemoteHost, rule.RemotePort)
	listener, err := sshClient.Listen("tcp", remoteAddr)
	if err != nil {
		return fmt.Errorf("failed to listen on remote %s: %v", remoteAddr, err)
	}

	session.listener = listener

	// Start accepting connections in a goroutine
	go func() {
		defer listener.Close()
		
		for {
			select {
			case <-session.done:
				return
			default:
				remoteConn, err := listener.Accept()
				if err != nil {
					if session.IsActive() {
						session.IncrementErrors(fmt.Sprintf("Remote accept error: %v", err))
					}
					continue
				}

				// Handle connection in separate goroutine
				go fm.handleRemoteForwardConnection(session, remoteConn, rule.LocalHost, rule.LocalPort)
			}
		}
	}()

	return nil
}

// handleRemoteForwardConnection handles a single remote forward connection
func (fm *ForwardingManager) handleRemoteForwardConnection(session *ForwardingSession, remoteConn net.Conn, localHost string, localPort int) {
	defer remoteConn.Close()
	
	session.IncrementConnections()
	defer session.DecrementActiveConnections()

	// Connect to local host
	localAddr := fmt.Sprintf("%s:%d", localHost, localPort)
	localConn, err := net.Dial("tcp", localAddr)
	if err != nil {
		session.IncrementErrors(fmt.Sprintf("Failed to connect to local %s: %v", localAddr, err))
		return
	}
	defer localConn.Close()

	// Start data forwarding
	fm.forwardData(session, localConn, remoteConn)
}

// startDynamicForwarding implements dynamic port forwarding (-D)
// Creates a SOCKS5 proxy on the local port
func (fm *ForwardingManager) startDynamicForwarding(session *ForwardingSession, host config.SSHHost, keyPassword string) error {
	rule := session.Rule
	
	// Get SSH client
	sshClient, err := fm.getSSHClient(host, keyPassword)
	if err != nil {
		return fmt.Errorf("failed to get SSH client: %v", err)
	}

	// Listen on local port for SOCKS5 connections
	localAddr := fmt.Sprintf("%s:%d", rule.LocalHost, rule.LocalPort)
	listener, err := net.Listen("tcp", localAddr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %v", localAddr, err)
	}

	session.listener = listener

	// Start accepting connections in a goroutine
	go func() {
		defer listener.Close()
		
		for {
			select {
			case <-session.done:
				return
			default:
				// Accept connection with timeout
				if tcpListener, ok := listener.(*net.TCPListener); ok {
					tcpListener.SetDeadline(time.Now().Add(1 * time.Second))
				}
				
				localConn, err := listener.Accept()
				if err != nil {
					if ne, ok := err.(net.Error); ok && ne.Timeout() {
						continue // Timeout is expected for graceful shutdown
					}
					if session.IsActive() {
						session.IncrementErrors(fmt.Sprintf("SOCKS accept error: %v", err))
					}
					continue
				}

				// Handle SOCKS5 connection in separate goroutine
				go fm.handleSOCKS5Connection(session, sshClient, localConn)
			}
		}
	}()

	return nil
}

// handleSOCKS5Connection handles a SOCKS5 proxy connection
func (fm *ForwardingManager) handleSOCKS5Connection(session *ForwardingSession, sshClient *ssh.Client, localConn net.Conn) {
	defer localConn.Close()
	
	session.IncrementConnections()
	defer session.DecrementActiveConnections()

	// Perform SOCKS5 handshake
	targetAddr, err := fm.socks5Handshake(localConn)
	if err != nil {
		session.IncrementErrors(fmt.Sprintf("SOCKS5 handshake failed: %v", err))
		return
	}

	// Connect to target through SSH
	remoteConn, err := sshClient.Dial("tcp", targetAddr)
	if err != nil {
		session.IncrementErrors(fmt.Sprintf("Failed to connect to %s: %v", targetAddr, err))
		// Send SOCKS5 error response
		localConn.Write([]byte{0x05, 0x05, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
		return
	}
	defer remoteConn.Close()

	// Send SOCKS5 success response
	localConn.Write([]byte{0x05, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})

	// Start data forwarding
	fm.forwardData(session, localConn, remoteConn)
}

// socks5Handshake performs SOCKS5 handshake and returns target address
func (fm *ForwardingManager) socks5Handshake(conn net.Conn) (string, error) {
	// Read initial request
	buf := make([]byte, 256)
	n, err := conn.Read(buf)
	if err != nil {
		return "", err
	}

	// Check SOCKS version
	if n < 3 || buf[0] != 0x05 {
		return "", fmt.Errorf("unsupported SOCKS version")
	}

	// Send auth method response (no auth required)
	conn.Write([]byte{0x05, 0x00})

	// Read connection request
	n, err = conn.Read(buf)
	if err != nil {
		return "", err
	}

	if n < 7 || buf[0] != 0x05 || buf[1] != 0x01 {
		return "", fmt.Errorf("invalid SOCKS5 request")
	}

	// Parse target address
	var targetAddr string
	switch buf[3] {
	case 0x01: // IPv4
		if n < 10 {
			return "", fmt.Errorf("invalid IPv4 address")
		}
		targetAddr = fmt.Sprintf("%d.%d.%d.%d:%d", buf[4], buf[5], buf[6], buf[7], int(buf[8])<<8+int(buf[9]))
	case 0x03: // Domain name
		if n < 7 {
			return "", fmt.Errorf("invalid domain name")
		}
		domainLen := int(buf[4])
		if n < 7+domainLen {
			return "", fmt.Errorf("incomplete domain name")
		}
		domain := string(buf[5 : 5+domainLen])
		port := int(buf[5+domainLen])<<8 + int(buf[6+domainLen])
		targetAddr = fmt.Sprintf("%s:%d", domain, port)
	default:
		return "", fmt.Errorf("unsupported address type")
	}

	return targetAddr, nil
}

// forwardData forwards data between two connections with statistics tracking
func (fm *ForwardingManager) forwardData(session *ForwardingSession, conn1, conn2 net.Conn) {
	done := make(chan struct{}, 2)

	// Forward conn1 -> conn2
	go func() {
		defer func() { done <- struct{}{} }()
		written, err := fm.copyWithStats(conn2, conn1, func(bytes int64) {
			session.AddBytesSent(bytes)
		})
		if err != nil && session.IsActive() {
			session.IncrementErrors(fmt.Sprintf("Forward error (sent %d bytes): %v", written, err))
		}
	}()

	// Forward conn2 -> conn1
	go func() {
		defer func() { done <- struct{}{} }()
		written, err := fm.copyWithStats(conn1, conn2, func(bytes int64) {
			session.AddBytesReceived(bytes)
		})
		if err != nil && session.IsActive() {
			session.IncrementErrors(fmt.Sprintf("Forward error (received %d bytes): %v", written, err))
		}
	}()

	// Wait for one direction to complete
	<-done
}

// copyWithStats copies data between connections while tracking statistics
func (fm *ForwardingManager) copyWithStats(dst, src net.Conn, statsCallback func(int64)) (int64, error) {
	buf := make([]byte, 32*1024) // 32KB buffer for better performance
	var written int64
	
	for {
		nr, er := src.Read(buf)
		if nr > 0 {
			nw, ew := dst.Write(buf[0:nr])
			if nw > 0 {
				written += int64(nw)
				statsCallback(int64(nw))
			}
			if ew != nil {
				return written, ew
			}
			if nr != nw {
				return written, io.ErrShortWrite
			}
		}
		if er != nil {
			if er != io.EOF {
				return written, er
			}
			break
		}
	}
	return written, nil
}