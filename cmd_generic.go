// Commands from http://redis.io/commands#generic

package rediqueue

import (
	"math/rand"
	"strconv"
	"strings"

	"github.com/chinahdkj/rediqueue/server"
)

// commandsGeneric handles EXPIRE, TTL, PERSIST, &c.
func commandsGeneric(m *RediQueue) {
	m.srv.Register("DEL", m.cmdDel)
	// DUMP
	m.srv.Register("EXISTS", m.cmdExists)
	m.srv.Register("KEYS", m.cmdKeys)
	// MIGRATE
	m.srv.Register("MOVE", m.cmdMove)
	// OBJECT
	m.srv.Register("RANDOMKEY", m.cmdRandomkey)
	m.srv.Register("RENAME", m.cmdRename)
	m.srv.Register("RENAMENX", m.cmdRenamenx)
	// RESTORE
	// SORT
	m.srv.Register("TYPE", m.cmdType)
	m.srv.Register("SCAN", m.cmdScan)
}

// DEL
func (m *RediQueue) cmdDel(c *server.Peer, cmd string, args []string) {
	if !m.handleAuth(c) {
		return
	}

	withTx(m, c, func(c *server.Peer, ctx *connCtx) {
		db := m.db(ctx.selectedDB)

		count := 0
		for _, key := range args {
			if db.exists(key) {
				count++
			}
			db.del(key) // delete expire
		}
		c.WriteInt(count)
	})
}

// TYPE
func (m *RediQueue) cmdType(c *server.Peer, cmd string, args []string) {
	if len(args) != 1 {
		setDirty(c)
		c.WriteError("usage error")
		return
	}
	if !m.handleAuth(c) {
		return
	}

	key := args[0]

	withTx(m, c, func(c *server.Peer, ctx *connCtx) {
		db := m.db(ctx.selectedDB)

		t, ok := db.keys[key]
		if !ok {
			c.WriteInline("none")
			return
		}

		c.WriteInline(t)
	})
}

// EXISTS
func (m *RediQueue) cmdExists(c *server.Peer, cmd string, args []string) {
	if len(args) < 1 {
		setDirty(c)
		c.WriteError(errWrongNumber(cmd))
		return
	}
	if !m.handleAuth(c) {
		return
	}

	withTx(m, c, func(c *server.Peer, ctx *connCtx) {
		db := m.db(ctx.selectedDB)

		found := 0
		for _, k := range args {
			if db.exists(k) {
				found++
			}
		}
		c.WriteInt(found)
	})
}

// MOVE
func (m *RediQueue) cmdMove(c *server.Peer, cmd string, args []string) {
	if len(args) != 2 {
		setDirty(c)
		c.WriteError(errWrongNumber(cmd))
		return
	}
	if !m.handleAuth(c) {
		return
	}

	key := args[0]
	targetDB, err := strconv.Atoi(args[1])
	if err != nil {
		targetDB = 0
	}

	withTx(m, c, func(c *server.Peer, ctx *connCtx) {
		if ctx.selectedDB == targetDB {
			c.WriteError("ERR source and destination objects are the same")
			return
		}
		db := m.db(ctx.selectedDB)
		targetDB := m.db(targetDB)

		if !db.move(key, targetDB) {
			c.WriteInt(0)
			return
		}
		c.WriteInt(1)
	})
}

// KEYS
func (m *RediQueue) cmdKeys(c *server.Peer, cmd string, args []string) {
	if len(args) != 1 {
		setDirty(c)
		c.WriteError(errWrongNumber(cmd))
		return
	}
	if !m.handleAuth(c) {
		return
	}

	key := args[0]

	withTx(m, c, func(c *server.Peer, ctx *connCtx) {
		db := m.db(ctx.selectedDB)

		keys := matchKeys(db.allKeys(), key)
		c.WriteLen(len(keys))
		for _, s := range keys {
			c.WriteBulk(s)
		}
	})
}

// RANDOMKEY
func (m *RediQueue) cmdRandomkey(c *server.Peer, cmd string, args []string) {
	if len(args) != 0 {
		setDirty(c)
		c.WriteError(errWrongNumber(cmd))
		return
	}
	if !m.handleAuth(c) {
		return
	}

	withTx(m, c, func(c *server.Peer, ctx *connCtx) {
		db := m.db(ctx.selectedDB)

		if len(db.keys) == 0 {
			c.WriteNull()
			return
		}
		nr := rand.Intn(len(db.keys))
		for k := range db.keys {
			if nr == 0 {
				c.WriteBulk(k)
				return
			}
			nr--
		}
	})
}

// RENAME
func (m *RediQueue) cmdRename(c *server.Peer, cmd string, args []string) {
	if len(args) != 2 {
		setDirty(c)
		c.WriteError(errWrongNumber(cmd))
		return
	}
	if !m.handleAuth(c) {
		return
	}

	from, to := args[0], args[1]

	withTx(m, c, func(c *server.Peer, ctx *connCtx) {
		db := m.db(ctx.selectedDB)

		if !db.exists(from) {
			c.WriteError(msgKeyNotFound)
			return
		}

		db.rename(from, to)
		c.WriteOK()
	})
}

// RENAMENX
func (m *RediQueue) cmdRenamenx(c *server.Peer, cmd string, args []string) {
	if len(args) != 2 {
		setDirty(c)
		c.WriteError(errWrongNumber(cmd))
		return
	}
	if !m.handleAuth(c) {
		return
	}

	from, to := args[0], args[1]

	withTx(m, c, func(c *server.Peer, ctx *connCtx) {
		db := m.db(ctx.selectedDB)

		if !db.exists(from) {
			c.WriteError(msgKeyNotFound)
			return
		}

		if db.exists(to) {
			c.WriteInt(0)
			return
		}

		db.rename(from, to)
		c.WriteInt(1)
	})
}

// SCAN
func (m *RediQueue) cmdScan(c *server.Peer, cmd string, args []string) {
	if len(args) < 1 {
		setDirty(c)
		c.WriteError(errWrongNumber(cmd))
		return
	}
	if !m.handleAuth(c) {
		return
	}

	cursor, err := strconv.Atoi(args[0])
	if err != nil {
		setDirty(c)
		c.WriteError(msgInvalidCursor)
		return
	}
	args = args[1:]

	// MATCH and COUNT options
	var withMatch bool
	var match string
	for len(args) > 0 {
		if strings.ToLower(args[0]) == "count" {
			// we do nothing with count
			if len(args) < 2 {
				setDirty(c)
				c.WriteError(msgSyntaxError)
				return
			}
			if _, err := strconv.Atoi(args[1]); err != nil {
				setDirty(c)
				c.WriteError(msgInvalidInt)
				return
			}
			args = args[2:]
			continue
		}
		if strings.ToLower(args[0]) == "match" {
			if len(args) < 2 {
				setDirty(c)
				c.WriteError(msgSyntaxError)
				return
			}
			withMatch = true
			match, args = args[1], args[2:]
			continue
		}
		setDirty(c)
		c.WriteError(msgSyntaxError)
		return
	}

	withTx(m, c, func(c *server.Peer, ctx *connCtx) {
		db := m.db(ctx.selectedDB)
		// We return _all_ (matched) keys every time.

		if cursor != 0 {
			// Invalid cursor.
			c.WriteLen(2)
			c.WriteBulk("0") // no next cursor
			c.WriteLen(0)    // no elements
			return
		}

		keys := db.allKeys()
		if withMatch {
			keys = matchKeys(keys, match)
		}

		c.WriteLen(2)
		c.WriteBulk("0") // no next cursor
		c.WriteLen(len(keys))
		for _, k := range keys {
			c.WriteBulk(k)
		}
	})
}
