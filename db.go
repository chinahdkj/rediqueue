package rediqueue

import (
	"sort"
)

func (db *RedisDB) exists(k string) bool {
	_, ok := db.keys[k]
	return ok
}

// t gives the type of a key, or ""
func (db *RedisDB) t(k string) string {
	return db.keys[k]
}

// allKeys returns all keys. Sorted.
func (db *RedisDB) allKeys() []string {
	res := make([]string, 0, len(db.keys))
	for k := range db.keys {
		res = append(res, k)
	}
	sort.Strings(res) // To make things deterministic.
	return res
}

// flush removes all keys and values.
func (db *RedisDB) flush() {
	db.keys = map[string]string{}
	db.listKeys = map[string]listKey{}
	db.setKeys = map[string]setKey{}
}

// move something to another db. Will return ok. Or not.
func (db *RedisDB) move(key string, to *RedisDB) bool {
	if _, ok := to.keys[key]; ok {
		return false
	}

	t, ok := db.keys[key]
	if !ok {
		return false
	}
	to.keys[key] = db.keys[key]
	switch t {
	case "list":
		to.listKeys[key] = db.listKeys[key]
	case "set":
		to.setKeys[key] = db.setKeys[key]
	default:
		panic("unhandled key type")
	}
	to.keyVersion[key]++
	db.del(key)
	return true
}

func (db *RedisDB) rename(from, to string) {
	db.del(to)
	switch db.t(from) {
	case "list":
		db.listKeys[to] = db.listKeys[from]
	case "set":
		db.setKeys[to] = db.setKeys[from]
	default:
		panic("missing case")
	}
	db.keys[to] = db.keys[from]
	db.keyVersion[to]++

	db.del(from)
}

func (db *RedisDB) del(k string) {
	if !db.exists(k) {
		return
	}
	t := db.t(k)
	delete(db.keys, k)
	db.keyVersion[k]++
	switch t {
	case "list":
		delete(db.listKeys, k)
	case "set":
		delete(db.setKeys, k)
	default:
		panic("Unknown key type: " + t)
	}
}

// listLpush is 'left push', aka unshift. Returns the new length.
func (db *RedisDB) listLpush(k, v string) int {
	l, ok := db.listKeys[k]
	if !ok {
		db.keys[k] = "list"
	}
	l = append([]string{v}, l...)
	db.listKeys[k] = l
	db.keyVersion[k]++
	return len(l)
}

// 'left pop', aka shift.
func (db *RedisDB) listLpop(k string) string {
	l := db.listKeys[k]
	el := l[0]
	l = l[1:]
	if len(l) == 0 {
		db.del(k)
	} else {
		db.listKeys[k] = l
	}
	db.keyVersion[k]++
	return el
}

func (db *RedisDB) listPush(k string, v ...string) int {
	l, ok := db.listKeys[k]
	if !ok {
		db.keys[k] = "list"
	}
	l = append(l, v...)
	db.listKeys[k] = l
	db.keyVersion[k]++
	return len(l)
}

func (db *RedisDB) listPop(k string) string {
	l := db.listKeys[k]
	el := l[len(l)-1]
	l = l[:len(l)-1]
	if len(l) == 0 {
		db.del(k)
	} else {
		db.listKeys[k] = l
		db.keyVersion[k]++
	}
	return el
}

// setset replaces a whole set.
func (db *RedisDB) setSet(k string, set setKey) {
	db.keys[k] = "set"
	db.setKeys[k] = set
	db.keyVersion[k]++
}

// setadd adds members to a set. Returns nr of new keys.
func (db *RedisDB) setAdd(k string, elems ...string) int {
	s, ok := db.setKeys[k]
	if !ok {
		s = setKey{}
		db.keys[k] = "set"
	}
	added := 0
	for _, e := range elems {
		if _, ok := s[e]; !ok {
			added++
		}
		s[e] = struct{}{}
	}
	db.setKeys[k] = s
	db.keyVersion[k]++
	return added
}

// setrem removes members from a set. Returns nr of deleted keys.
func (db *RedisDB) setRem(k string, fields ...string) int {
	s, ok := db.setKeys[k]
	if !ok {
		return 0
	}
	removed := 0
	for _, f := range fields {
		if _, ok := s[f]; ok {
			removed++
			delete(s, f)
		}
	}
	if len(s) == 0 {
		db.del(k)
	} else {
		db.setKeys[k] = s
	}
	db.keyVersion[k]++
	return removed
}

// All members of a set.
func (db *RedisDB) setMembers(k string) []string {
	set := db.setKeys[k]
	members := make([]string, 0, len(set))
	for k := range set {
		members = append(members, k)
	}
	sort.Strings(members)
	return members
}

// Is a SET value present?
func (db *RedisDB) setIsMember(k, v string) bool {
	set, ok := db.setKeys[k]
	if !ok {
		return false
	}
	_, ok = set[v]
	return ok
}

// setDiff implements the logic behind SDIFF*
func (db *RedisDB) setDiff(keys []string) (setKey, error) {
	key := keys[0]
	keys = keys[1:]
	if db.exists(key) && db.t(key) != "set" {
		return nil, ErrWrongType
	}
	s := setKey{}
	for k := range db.setKeys[key] {
		s[k] = struct{}{}
	}
	for _, sk := range keys {
		if !db.exists(sk) {
			continue
		}
		if db.t(sk) != "set" {
			return nil, ErrWrongType
		}
		for e := range db.setKeys[sk] {
			delete(s, e)
		}
	}
	return s, nil
}

// setInter implements the logic behind SINTER*
func (db *RedisDB) setInter(keys []string) (setKey, error) {
	key := keys[0]
	keys = keys[1:]
	if db.exists(key) && db.t(key) != "set" {
		return nil, ErrWrongType
	}
	s := setKey{}
	for k := range db.setKeys[key] {
		s[k] = struct{}{}
	}
	for _, sk := range keys {
		if !db.exists(sk) {
			continue
		}
		if db.t(sk) != "set" {
			// Bug(?) in redis 2.8.14, it just skips the key.
			continue
			// return nil, ErrWrongType
		}
		other := db.setKeys[sk]
		for e := range s {
			if _, ok := other[e]; ok {
				continue
			}
			delete(s, e)
		}
	}
	return s, nil
}

// setUnion implements the logic behind SUNION*
func (db *RedisDB) setUnion(keys []string) (setKey, error) {
	key := keys[0]
	keys = keys[1:]
	if db.exists(key) && db.t(key) != "set" {
		return nil, ErrWrongType
	}
	s := setKey{}
	for k := range db.setKeys[key] {
		s[k] = struct{}{}
	}
	for _, sk := range keys {
		if !db.exists(sk) {
			continue
		}
		if db.t(sk) != "set" {
			return nil, ErrWrongType
		}
		for e := range db.setKeys[sk] {
			s[e] = struct{}{}
		}
	}
	return s, nil
}

func reverseSlice(o []string) {
	for i := range make([]struct{}, len(o)/2) {
		other := len(o) - 1 - i
		o[i], o[other] = o[other], o[i]
	}
}
