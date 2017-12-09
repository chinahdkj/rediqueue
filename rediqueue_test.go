package rediqueue

import (
	"testing"

	"github.com/garyburd/redigo/redis"
)

// Test starting/stopping a server
func TestServer(t *testing.T) {
	s, err := Run()
	ok(t, err)
	defer s.Close()

	c, err := redis.Dial("tcp", s.Addr())
	ok(t, err)
	_, err = c.Do("PING")
	ok(t, err)

	// A single client
	equals(t, 1, s.CurrentConnectionCount())
	equals(t, 1, s.TotalConnectionCount())
	equals(t, 1, s.CommandCount())
	_, err = c.Do("PING")
	ok(t, err)
	equals(t, 2, s.CommandCount())
}

func TestMultipleServers(t *testing.T) {
	s1, err := Run()
	ok(t, err)
	s2, err := Run()
	ok(t, err)
	if s1.Addr() == s2.Addr() {
		t.Fatal("Non-unique addresses", s1.Addr(), s2.Addr())
	}

	s2.Close()
	s1.Close()
	// Closing multiple times is fine
	go s1.Close()
	go s1.Close()
	s1.Close()
}

func TestRestart(t *testing.T) {
	s, err := Run()
	ok(t, err)
	addr := s.Addr()

	s.Lpush("color", "red")

	s.Close()
	err = s.Restart()
	ok(t, err)
	if have, want := s.Addr(), addr; have != want {
		t.Fatalf("have: %s, want: %s", have, want)
	}

	c, err := redis.Dial("tcp", s.Addr())
	ok(t, err)
	_, err = c.Do("PING")
	ok(t, err)

	red, err := redis.String(c.Do("LPOP", "color"))
	ok(t, err)
	if have, want := red, "red"; have != want {
		t.Errorf("have: %s, want: %s", have, want)
	}
}

// Test a custom addr
func TestAddr(t *testing.T) {
	m := NewRediQueue()
	err := m.StartAddr("127.0.0.1:7887")
	ok(t, err)
	defer m.Close()

	c, err := redis.Dial("tcp", "127.0.0.1:7887")
	ok(t, err)
	_, err = c.Do("PING")
	ok(t, err)
}

func TestDumpList(t *testing.T) {
	s, err := Run()
	ok(t, err)
	s.Push("elements", "earth")
	s.Push("elements", "wind")
	s.Push("elements", "fire")
	if have, want := s.Dump(), `- elements
   "earth"
   "wind"
   "fire"
`; have != want {
		t.Errorf("have: %q, want: %q", have, want)
	}
}

func TestDumpSet(t *testing.T) {
	s, err := Run()
	ok(t, err)
	s.SetAdd("elements", "earth")
	s.SetAdd("elements", "wind")
	s.SetAdd("elements", "fire")
	if have, want := s.Dump(), `- elements
   "earth"
   "fire"
   "wind"
`; have != want {
		t.Errorf("have: %q, want: %q", have, want)
	}
}
