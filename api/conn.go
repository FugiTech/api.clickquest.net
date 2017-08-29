package api

import (
	"sync"
	"time"
)

var (
	connected []*Conn
	connLock  sync.RWMutex
)

type Conn struct {
	out chan *messageOut
	IP  string

	sync.Mutex
	User          *User
	LastHeartbeat time.Time
	LastActivity  time.Time
}

func NewConn(ip string) *Conn {
	connLock.Lock()
	defer connLock.Unlock()
	c := &Conn{
		out: make(chan *messageOut, 10),
		IP:  ip,
	}
	connected = append(connected, c)
	return c
}

func (c *Conn) Disconnect() {
	defer broadcastPlayers() // So it runs after releasing the lock
	connLock.Lock()
	defer connLock.Unlock()

	close(c.out)
	i := 0
	for ; i < len(connected); i++ {
		if connected[i] == c {
			break
		}
	}
	connected = append(connected[:i], connected[i+1:]...)
}

func Broadcast(typ string, data interface{}) {
	connLock.RLock()
	defer connLock.RUnlock()
	for _, conn := range connected {
		m := &messageOut{
			Type: typ,
			Data: data,
		}
		select {
		case conn.out <- m:
		default:
		}
	}
}

func broadcastPlayers() {
	Broadcast("players", getPlayers())
}

func getPlayers() []User {
	connLock.RLock()
	defer connLock.RUnlock()
	users := []User{}
	for _, conn := range connected {
		conn.Lock()
		if conn.User != nil {
			users = append(users, *conn.User)
		}
		conn.Unlock()
	}
	return users
}
