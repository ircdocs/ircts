// Copyright (c) 2017 Daniel Oaks <daniel@danieloaks.net>
// released under the ISC license

package conn

import (
	"errors"
	"sync"

	"github.com/goshuirc/irc-go/ircmsg"
	"github.com/ircdocs/ircts/lib/utils"
)

// Connection holds info about a single IRC connection.
type Connection struct {
	nick       string
	socket     *Socket
	traffic    string
	stateMutex sync.RWMutex
}

// Nick returns the connection's current nickname.
func (c *Connection) Nick() string {
	c.stateMutex.RLock()
	defer c.stateMutex.RUnlock()
	return c.nick
}

// Traffic returns the traffic that has gone past on the connection.
func (c *Connection) Traffic() string {
	c.stateMutex.RLock()
	defer c.stateMutex.RUnlock()
	return c.traffic
}

// SendLine sends a line to the server.
func (c *Connection) SendLine(line string) error {
	c.stateMutex.Lock()
	defer c.stateMutex.Unlock()
	c.traffic += " -> " + line + "\n"
	return c.socket.SendLine(line)
}

// SendMessage sends an IRC message to the server.
func (c *Connection) SendMessage(tags *map[string]ircmsg.TagValue, prefix, command string, params ...string) error {
	return c.socket.SendMessage(tags, prefix, command, params...)
}

// SendSimpleMessage sends a simple IRC message to the server.
func (c *Connection) SendSimpleMessage(command string, params ...string) error {
	return c.SendMessage(nil, "", command, params...)
}

// GetLine returns a single IRC line from the server.
func (c *Connection) GetLine() (string, error) {
	line, err := c.socket.GetLine()
	if err == nil {
		c.stateMutex.Lock()
		defer c.stateMutex.Unlock()
		c.traffic += "<-  " + line + "\n"
	}
	return line, err
}

// ConnectionPool holds a bunch of connections, and helps simplify handling of
// the ircts config around optionally resetting connections.
type ConnectionPool struct {
	// true/false for in use or not in use
	connections      map[*Connection]bool
	connectionsMutex sync.Mutex
}

// NewConnectionPool returns a new ConnectionPool.
func NewConnectionPool() *ConnectionPool {
	var cp ConnectionPool
	cp.connections = make(map[*Connection]bool)
	return &cp
}

// NewConnection returns an entirely new connection from our pool.
func (cp *ConnectionPool) NewConnection(sc utils.ServerConfig) (*Connection, error) {
	socket, err := ConnectSocket(sc.Address, sc.TLS, insecureTLSConfig)
	if err != nil {
		return nil, errors.New("Failed to connect to server: " + err.Error())
	}

	c := Connection{
		socket: socket,
	}

	cp.connectionsMutex.Lock()
	defer cp.connectionsMutex.Unlock()
	cp.connections[&c] = true

	return &c, nil
}

// DestroyConnection marks the given connection as destroyed and removes it from our list.
func (cp *ConnectionPool) DestroyConnection(c *Connection) {
	// disconnect
	c.socket.Disconnect()

	// remove from our connection list
	cp.connectionsMutex.Lock()
	defer cp.connectionsMutex.Unlock()
	delete(cp.connections, c)
}
