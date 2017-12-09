package rediqueue

// Commands to modify and query our databases directly.

import (
	"errors"
)

var (
	// ErrKeyNotFound is returned when a key doesn't exist.
	ErrKeyNotFound = errors.New(msgKeyNotFound)
	// ErrWrongType when a key is not the right type.
	ErrWrongType = errors.New(msgWrongType)
	// ErrIntValueError can returned by INCRBY
	ErrIntValueError = errors.New(msgInvalidInt)
	// ErrFloatValueError can returned by INCRBYFLOAT
	ErrFloatValueError = errors.New(msgInvalidFloat)
)

// Select sets the DB id for all direct commands.
func (m *RediQueue) Select(i int) {
	m.Lock()
	defer m.Unlock()
	m.selectedDB = i
}

// Keys returns all keys from the selected database, sorted.
func (m *RediQueue) Keys() []string {
	return m.DB(m.selectedDB).Keys()
}

// Keys returns all keys, sorted.
func (db *RedisDB) Keys() []string {
	db.master.Lock()
	defer db.master.Unlock()
	return db.allKeys()
}

// FlushAll removes all keys from all databases.
func (m *RediQueue) FlushAll() {
	m.Lock()
	defer m.Unlock()
	m.flushAll()
}

func (m *RediQueue) flushAll() {
	for _, db := range m.dbs {
		db.flush()
	}
}

// FlushDB removes all keys from the selected database.
func (m *RediQueue) FlushDB() {
	m.DB(m.selectedDB).FlushDB()
}

// FlushDB removes all keys.
func (db *RedisDB) FlushDB() {
	db.master.Lock()
	defer db.master.Unlock()
	db.flush()
}

// List returns the list k, or an error if it's not there or something else.
// This is the same as the Redis command `LRANGE 0 -1`, but you can do your own
// range-ing.
func (m *RediQueue) List(k string) ([]string, error) {
	return m.DB(m.selectedDB).List(k)
}

// List returns the list k, or an error if it's not there or something else.
// This is the same as the Redis command `LRANGE 0 -1`, but you can do your own
// range-ing.
func (db *RedisDB) List(k string) ([]string, error) {
	db.master.Lock()
	defer db.master.Unlock()

	if !db.exists(k) {
		return nil, ErrKeyNotFound
	}
	if db.t(k) != "list" {
		return nil, ErrWrongType
	}
	return db.listKeys[k], nil
}

// Lpush is an unshift. Returns the new length.
func (m *RediQueue) Lpush(k, v string) (int, error) {
	return m.DB(m.selectedDB).Lpush(k, v)
}

// Lpush is an unshift. Returns the new length.
func (db *RedisDB) Lpush(k, v string) (int, error) {
	db.master.Lock()
	defer db.master.Unlock()

	if db.exists(k) && db.t(k) != "list" {
		return 0, ErrWrongType
	}
	return db.listLpush(k, v), nil
}

// Lpop is a shift. Returns the popped element.
func (m *RediQueue) Lpop(k string) (string, error) {
	return m.DB(m.selectedDB).Lpop(k)
}

// Lpop is a shift. Returns the popped element.
func (db *RedisDB) Lpop(k string) (string, error) {
	db.master.Lock()
	defer db.master.Unlock()

	if !db.exists(k) {
		return "", ErrKeyNotFound
	}
	if db.t(k) != "list" {
		return "", ErrWrongType
	}
	return db.listLpop(k), nil
}

// Push add element at the end. Is called RPUSH in redis. Returns the new length.
func (m *RediQueue) Push(k string, v ...string) (int, error) {
	return m.DB(m.selectedDB).Push(k, v...)
}

// Push add element at the end. Is called RPUSH in redis. Returns the new length.
func (db *RedisDB) Push(k string, v ...string) (int, error) {
	db.master.Lock()
	defer db.master.Unlock()

	if db.exists(k) && db.t(k) != "list" {
		return 0, ErrWrongType
	}
	return db.listPush(k, v...), nil
}

// Pop removes and returns the last element. Is called RPOP in Redis.
func (m *RediQueue) Pop(k string) (string, error) {
	return m.DB(m.selectedDB).Pop(k)
}

// Pop removes and returns the last element. Is called RPOP in Redis.
func (db *RedisDB) Pop(k string) (string, error) {
	db.master.Lock()
	defer db.master.Unlock()

	if !db.exists(k) {
		return "", ErrKeyNotFound
	}
	if db.t(k) != "list" {
		return "", ErrWrongType
	}

	return db.listPop(k), nil
}

// SetAdd adds keys to a set. Returns the number of new keys.
func (m *RediQueue) SetAdd(k string, elems ...string) (int, error) {
	return m.DB(m.selectedDB).SetAdd(k, elems...)
}

// SetAdd adds keys to a set. Returns the number of new keys.
func (db *RedisDB) SetAdd(k string, elems ...string) (int, error) {
	db.master.Lock()
	defer db.master.Unlock()
	if db.exists(k) && db.t(k) != "set" {
		return 0, ErrWrongType
	}
	return db.setAdd(k, elems...), nil
}

// Members gives all set keys. Sorted.
func (m *RediQueue) Members(k string) ([]string, error) {
	return m.DB(m.selectedDB).Members(k)
}

// Members gives all set keys. Sorted.
func (db *RedisDB) Members(k string) ([]string, error) {
	db.master.Lock()
	defer db.master.Unlock()
	if !db.exists(k) {
		return nil, ErrKeyNotFound
	}
	if db.t(k) != "set" {
		return nil, ErrWrongType
	}
	return db.setMembers(k), nil
}

// IsMember tells if value is in the set.
func (m *RediQueue) IsMember(k, v string) (bool, error) {
	return m.DB(m.selectedDB).IsMember(k, v)
}

// IsMember tells if value is in the set.
func (db *RedisDB) IsMember(k, v string) (bool, error) {
	db.master.Lock()
	defer db.master.Unlock()
	if !db.exists(k) {
		return false, ErrKeyNotFound
	}
	if db.t(k) != "set" {
		return false, ErrWrongType
	}
	return db.setIsMember(k, v), nil
}

// Del deletes a key and any expiration value. Returns whether there was a key.
func (m *RediQueue) Del(k string) bool {
	return m.DB(m.selectedDB).Del(k)
}

// Del deletes a key and any expiration value. Returns whether there was a key.
func (db *RedisDB) Del(k string) bool {
	db.master.Lock()
	defer db.master.Unlock()
	if !db.exists(k) {
		return false
	}
	db.del(k)
	return true
}

// Type gives the type of a key, or ""
func (m *RediQueue) Type(k string) string {
	return m.DB(m.selectedDB).Type(k)
}

// Type gives the type of a key, or ""
func (db *RedisDB) Type(k string) string {
	db.master.Lock()
	defer db.master.Unlock()
	return db.t(k)
}

// Exists tells whether a key exists.
func (m *RediQueue) Exists(k string) bool {
	return m.DB(m.selectedDB).Exists(k)
}

// Exists tells whether a key exists.
func (db *RedisDB) Exists(k string) bool {
	db.master.Lock()
	defer db.master.Unlock()
	return db.exists(k)
}

// SRem removes fields from a set. Returns number of deleted fields.
func (m *RediQueue) SRem(k string, fields ...string) (int, error) {
	return m.DB(m.selectedDB).SRem(k, fields...)
}

// SRem removes fields from a set. Returns number of deleted fields.
func (db *RedisDB) SRem(k string, fields ...string) (int, error) {
	db.master.Lock()
	defer db.master.Unlock()
	if !db.exists(k) {
		return 0, ErrKeyNotFound
	}
	if db.t(k) != "set" {
		return 0, ErrWrongType
	}
	return db.setRem(k, fields...), nil
}
