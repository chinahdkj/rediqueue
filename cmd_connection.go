// Commands from http://redis.io/commands#connection

package rediqueue

import (
	"strconv"

	"github.com/chinahdkj/rediqueue/server"
)

func commandsConnection(m *RediQueue) {
	m.srv.Register("AUTH", m.cmdAuth)
	m.srv.Register("ECHO", m.cmdEcho)
	m.srv.Register("PING", m.cmdPing)
	m.srv.Register("SELECT", m.cmdSelect)
	m.srv.Register("QUIT", m.cmdQuit)
}

// PING
func (m *RediQueue) cmdPing(c *server.Peer, cmd string, args []string) {
	if !m.handleAuth(c) {
		return
	}
	c.WriteInline("PONG")
}

// AUTH
func (m *RediQueue) cmdAuth(c *server.Peer, cmd string, args []string) {
	if len(args) != 1 {
		setDirty(c)
		c.WriteError(errWrongNumber(cmd))
		return
	}
	pw := args[0]

	m.Lock()
	defer m.Unlock()
	if m.password == "" {
		c.WriteError("ERR Client sent AUTH, but no password is set")
		return
	}
	if m.password != pw {
		c.WriteError("ERR invalid password")
		return
	}

	setAuthenticated(c)
	c.WriteOK()
}

// ECHO
func (m *RediQueue) cmdEcho(c *server.Peer, cmd string, args []string) {
	if len(args) != 1 {
		setDirty(c)
		c.WriteError(errWrongNumber(cmd))
		return
	}
	if !m.handleAuth(c) {
		return
	}

	msg := args[0]
	c.WriteBulk(msg)
}

// SELECT
func (m *RediQueue) cmdSelect(c *server.Peer, cmd string, args []string) {
	if len(args) != 1 {
		setDirty(c)
		c.WriteError(errWrongNumber(cmd))
		return
	}
	if !m.handleAuth(c) {
		return
	}

	id, err := strconv.Atoi(args[0])
	if err != nil {
		id = 0
	}

	m.Lock()
	defer m.Unlock()

	ctx := getCtx(c)
	ctx.selectedDB = id

	c.WriteOK()
}

// QUIT
func (m *RediQueue) cmdQuit(c *server.Peer, cmd string, args []string) {
	// QUIT isn't transactionfied and accepts any arguments.
	c.WriteOK()
	c.Close()
}
