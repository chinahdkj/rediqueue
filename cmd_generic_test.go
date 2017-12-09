package rediqueue

import (
	"testing"

	"github.com/garyburd/redigo/redis"
)

func TestDel(t *testing.T) {

	s, err := Run()
	ok(t, err)

	defer s.Close()

	s.Lpush("foo", "123")
	s.SetAdd("bar", "123", "234")

	// Direct also works:
	r := s.Del("foo")
	equals(t, true, r)
}

func TestType(t *testing.T) {
	s, err := Run()
	ok(t, err)
	defer s.Close()
	c, err := redis.Dial("tcp", s.Addr())
	ok(t, err)

	// New key
	{
		v, err := redis.String(c.Do("TYPE", "nosuch"))
		ok(t, err)
		equals(t, "none", v)
	}

	// Wrong usage
	{
		_, err := redis.Int(c.Do("TYPE"))
		assert(t, err != nil, "do TYPE error")
		_, err = redis.Int(c.Do("TYPE", "spurious", "arguments"))
		assert(t, err != nil, "do TYPE error")
	}

	// Direct usage:
	{
		redis.Int(c.Do("LPUSH", "aap", "123"))
		equals(t, "list", s.Type("aap"))
		equals(t, "", s.Type("nokey"))
	}
}

func TestExists(t *testing.T) {
	s, err := Run()
	ok(t, err)
	defer s.Close()
	c, err := redis.Dial("tcp", s.Addr())
	ok(t, err)

	// String key
	{
		s.Lpush("foo", "1")
		v, err := redis.Int(c.Do("EXISTS", "foo"))
		ok(t, err)
		equals(t, 1, v)
	}

	// Hash key
	{
		s.SetAdd("aap", "2")
		v, err := redis.Int(c.Do("EXISTS", "aap"))
		ok(t, err)
		equals(t, 1, v)
	}

	// Multiple keys
	{
		v, err := redis.Int(c.Do("EXISTS", "foo", "aap"))
		ok(t, err)
		equals(t, 2, v)

		v, err = redis.Int(c.Do("EXISTS", "foo", "noot", "aap"))
		ok(t, err)
		equals(t, 2, v)
	}

	// New key
	{
		v, err := redis.Int(c.Do("EXISTS", "nosuch"))
		ok(t, err)
		equals(t, 0, v)
	}

	// Wrong usage
	{
		_, err := redis.Int(c.Do("EXISTS"))
		assert(t, err != nil, "do EXISTS error")
	}

	// Direct usage:
	{
		equals(t, true, s.Exists("aap"))
		equals(t, false, s.Exists("nokey"))
	}
}

func TestMove(t *testing.T) {
	s, err := Run()
	ok(t, err)
	defer s.Close()
	c, err := redis.Dial("tcp", s.Addr())
	ok(t, err)

	// No problem.
	{
		s.Lpush("foo", "bar!")
		v, err := redis.Int(c.Do("MOVE", "foo", 1))
		ok(t, err)
		equals(t, 1, v)
	}

	// Src key doesn't exists.
	{
		v, err := redis.Int(c.Do("MOVE", "nosuch", 1))
		ok(t, err)
		equals(t, 0, v)
	}

	// Target key already exists.
	{
		s.DB(0).Lpush("two", "orig")
		s.DB(1).Lpush("two", "taken")
		v, err := redis.Int(c.Do("MOVE", "two", 1))
		ok(t, err)
		equals(t, 0, v)
		s.CheckList(t, "two", "orig")
	}

	// TTL is also moved
	{
		s.DB(0).Lpush("one", "two")
		v, err := redis.Int(c.Do("MOVE", "one", 1))
		ok(t, err)
		equals(t, 1, v)
	}

	// Wrong usage
	{
		_, err := redis.Int(c.Do("MOVE"))
		assert(t, err != nil, "do MOVE error")
		_, err = redis.Int(c.Do("MOVE", "foo"))
		assert(t, err != nil, "do MOVE error")
		_, err = redis.Int(c.Do("MOVE", "foo", "noint"))
		assert(t, err != nil, "do MOVE error")
		_, err = redis.Int(c.Do("MOVE", "foo", 2, "toomany"))
		assert(t, err != nil, "do MOVE error")
	}
}

func TestKeys(t *testing.T) {
	s, err := Run()
	ok(t, err)
	defer s.Close()
	c, err := redis.Dial("tcp", s.Addr())
	ok(t, err)

	s.Lpush("foo", "bar!")
	s.Lpush("foobar", "bar!")
	s.Lpush("barfoo", "bar!")
	s.Lpush("fooooo", "bar!")

	{
		v, err := redis.Strings(c.Do("KEYS", "foo"))
		ok(t, err)
		equals(t, []string{"foo"}, v)
	}

	// simple '*'
	{
		v, err := redis.Strings(c.Do("KEYS", "foo*"))
		ok(t, err)
		equals(t, []string{"foo", "foobar", "fooooo"}, v)
	}
	// simple '?'
	{
		v, err := redis.Strings(c.Do("KEYS", "fo?"))
		ok(t, err)
		equals(t, []string{"foo"}, v)
	}

	// Don't die on never-matching pattern.
	{
		v, err := redis.Strings(c.Do("KEYS", `f\`))
		ok(t, err)
		equals(t, []string{}, v)
	}

	// Wrong usage
	{
		_, err := redis.Int(c.Do("KEYS"))
		assert(t, err != nil, "do KEYS error")
		_, err = redis.Int(c.Do("KEYS", "foo", "noint"))
		assert(t, err != nil, "do KEYS error")
	}
}

func TestRandom(t *testing.T) {
	s, err := Run()
	ok(t, err)
	defer s.Close()
	c, err := redis.Dial("tcp", s.Addr())
	ok(t, err)

	// Empty db.
	{
		v, err := c.Do("RANDOMKEY")
		ok(t, err)
		equals(t, nil, v)
	}

	s.Lpush("one", "bar!")
	s.Lpush("two", "bar!")
	s.Lpush("three", "bar!")

	// No idea which key will be returned.
	{
		v, err := redis.String(c.Do("RANDOMKEY"))
		ok(t, err)
		assert(t, v == "one" || v == "two" || v == "three", "RANDOMKEY looks sane")
	}

	// Wrong usage
	{
		_, err = redis.Int(c.Do("RANDOMKEY", "spurious"))
		assert(t, err != nil, "do RANDOMKEY error")
	}
}

func TestRename(t *testing.T) {
	s, err := Run()
	ok(t, err)
	defer s.Close()
	c, err := redis.Dial("tcp", s.Addr())
	ok(t, err)

	// Non-existing key
	{
		_, err := redis.Int(c.Do("RENAME", "nosuch", "to"))
		assert(t, err != nil, "do RENAME error")
	}

	// Same key
	{
		_, err := redis.Int(c.Do("RENAME", "from", "from"))
		assert(t, err != nil, "do RENAME error")
	}

	// Move a string key
	{
		s.Lpush("from", "value")
		str, err := redis.String(c.Do("RENAME", "from", "to"))
		ok(t, err)
		equals(t, "OK", str)
		equals(t, false, s.Exists("from"))
		equals(t, true, s.Exists("to"))
		s.CheckList(t, "to", "value")
	}

	// Wrong usage
	{
		_, err := redis.Int(c.Do("RENAME"))
		assert(t, err != nil, "do RENAME error")
		_, err = redis.Int(c.Do("RENAME", "too few"))
		assert(t, err != nil, "do RENAME error")
		_, err = redis.Int(c.Do("RENAME", "some", "spurious", "arguments"))
		assert(t, err != nil, "do RENAME error")
	}
}

func TestScan(t *testing.T) {
	s, err := Run()
	ok(t, err)
	defer s.Close()
	c, err := redis.Dial("tcp", s.Addr())
	ok(t, err)

	// We cheat with scan. It always returns everything.

	s.Lpush("key", "value")

	// No problem
	{
		res, err := redis.Values(c.Do("SCAN", 0))
		ok(t, err)
		equals(t, 2, len(res))

		var c int
		var keys []string
		_, err = redis.Scan(res, &c, &keys)
		ok(t, err)
		equals(t, 0, c)
		equals(t, []string{"key"}, keys)
	}

	// Invalid cursor
	{
		res, err := redis.Values(c.Do("SCAN", 42))
		ok(t, err)
		equals(t, 2, len(res))

		var c int
		var keys []string
		_, err = redis.Scan(res, &c, &keys)
		ok(t, err)
		equals(t, 0, c)
		equals(t, []string(nil), keys)
	}

	// COUNT (ignored)
	{
		res, err := redis.Values(c.Do("SCAN", 0, "COUNT", 200))
		ok(t, err)
		equals(t, 2, len(res))

		var c int
		var keys []string
		_, err = redis.Scan(res, &c, &keys)
		ok(t, err)
		equals(t, 0, c)
		equals(t, []string{"key"}, keys)
	}

	// MATCH
	{
		s.Lpush("aap", "noot")
		s.Lpush("mies", "wim")
		res, err := redis.Values(c.Do("SCAN", 0, "MATCH", "mi*"))
		ok(t, err)
		equals(t, 2, len(res))

		var c int
		var keys []string
		_, err = redis.Scan(res, &c, &keys)
		ok(t, err)
		equals(t, 0, c)
		equals(t, []string{"mies"}, keys)
	}

	// Wrong usage
	{
		_, err := redis.Int(c.Do("SCAN"))
		assert(t, err != nil, "do SCAN error")
		_, err = redis.Int(c.Do("SCAN", "noint"))
		assert(t, err != nil, "do SCAN error")
		_, err = redis.Int(c.Do("SCAN", 1, "MATCH"))
		assert(t, err != nil, "do SCAN error")
		_, err = redis.Int(c.Do("SCAN", 1, "COUNT"))
		assert(t, err != nil, "do SCAN error")
		_, err = redis.Int(c.Do("SCAN", 1, "COUNT", "noint"))
		assert(t, err != nil, "do SCAN error")
	}
}

func TestRenamenx(t *testing.T) {
	s, err := Run()
	ok(t, err)
	defer s.Close()
	c, err := redis.Dial("tcp", s.Addr())
	ok(t, err)

	// Non-existing key
	{
		_, err := redis.Int(c.Do("RENAMENX", "nosuch", "to"))
		assert(t, err != nil, "do RENAMENX error")
	}

	// Same key
	{
		_, err := redis.Int(c.Do("RENAMENX", "from", "from"))
		assert(t, err != nil, "do RENAMENX error")
	}

	// Move a string key
	{
		s.Lpush("from", "value")
		n, err := redis.Int(c.Do("RENAMENX", "from", "to"))
		ok(t, err)
		equals(t, 1, n)
		equals(t, false, s.Exists("from"))
		equals(t, true, s.Exists("to"))
		s.CheckList(t, "to", "value")
		c.Do("DEL", "to")
	}

	// Move over something which exists
	{
		s.Lpush("from", "string value")
		s.Lpush("to", "value")

		n, err := redis.Int(c.Do("RENAMENX", "from", "to"))
		ok(t, err)
		equals(t, 0, n)
		equals(t, true, s.Exists("from"))
		equals(t, true, s.Exists("to"))

		s.CheckList(t, "from", "string value")
		s.CheckList(t, "to", "value")
	}

	// Wrong usage
	{
		_, err := redis.Int(c.Do("RENAMENX"))
		assert(t, err != nil, "do RENAMENX error")
		_, err = redis.Int(c.Do("RENAMENX", "too few"))
		assert(t, err != nil, "do RENAMENX error")
		_, err = redis.Int(c.Do("RENAMENX", "some", "spurious", "arguments"))
		assert(t, err != nil, "do RENAMENX error")
	}
}
