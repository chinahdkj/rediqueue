package rediqueue_test

import (
	"github.com/chinahdkj/rediqueue"
	"github.com/garyburd/redigo/redis"
)

func Example() {
	s, err := rediqueue.Run()
	if err != nil {
		panic(err)
	}
	defer s.Close()

	// Configure you application to connect to redis at s.Addr()
	// Any redis client should work, as long as you use redis commands which
	// rediqueue implements.
	c, err := redis.Dial("tcp", s.Addr())
	if err != nil {
		panic(err)
	}
	if _, err = c.Do("LPUSH", "foo", "bar"); err != nil {
		panic(err)
	}

	// You can ask rediqueue about keys directly, without going over the network.
	if got, err := s.Lpop("foo"); err != nil || got != "bar" {
		panic("Didn't get 'bar' back")
	}
	// Or with a DB id
	if _, err := s.DB(42).Lpop("foo"); err != rediqueue.ErrKeyNotFound {
		panic("didn't use a different database")
	}

	// Or use a Check* function which Fail()s if the key is not what we expect
	// (checks for existence, key type and the value)
	// s.CheckGet(t, "foo", "bar")

	// Check if there really was only one connection.
	if s.TotalConnectionCount() != 1 {
		panic("too many connections made")
	}

	// Output:
}
