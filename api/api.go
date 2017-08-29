package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var (
	DB          *sql.DB
	latestStats *Stats
	LatestChat  []ChatLine
	Mutex       sync.Mutex
)

const LinesPerPage = 50

var ErrDisconnect = errors.New("disconnect")

type messageIn struct {
	ID   string
	Type string
	Data json.RawMessage
}

type messageOut struct {
	ID   string
	Type string
	Data interface{}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func API(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	ip := strings.Split(r.Header.Get("X-Forwarded-For"), ",")[0]
	if ip == "" {
		ip, _, _ = net.SplitHostPort(r.RemoteAddr)
	}

	Mutex.Lock()
	conn.WriteJSON(&messageOut{
		Type: "stats",
		Data: latestStats,
	})

	conn.WriteJSON(&messageOut{
		Type: "chat",
		Data: LatestChat,
	})
	Mutex.Unlock()

	conn.WriteJSON(&messageOut{
		Type: "players",
		Data: getPlayers(),
	})

	c := NewConn(ip)
	go func() {
		defer c.Disconnect()
		for {
			err := conn.SetReadDeadline(time.Now().Add(time.Minute))
			if err != nil {
				log.Println("conn.SetReadDeadline:", err)
				return
			}

			m := &messageIn{}
			err = conn.ReadJSON(&m)
			if err != nil {
				if !websocket.IsCloseError(err, websocket.CloseGoingAway) {
					log.Println("conn.ReadJSON:", err)
				}
				return
			}

			h, ok := handlers[m.Type]
			if !ok {
				c.out <- &messageOut{
					ID: m.ID,
					Data: map[string]interface{}{
						"error": fmt.Sprintf("Invalid message type: %s", m.Type),
					},
				}
				continue
			}

			var in interface{}
			if h.Params != nil {
				in = reflect.New(h.Params).Interface()
				err = json.Unmarshal([]byte(m.Data), &in)
				if err != nil {
					c.out <- &messageOut{
						ID: m.ID,
						Data: map[string]interface{}{
							"error": err.Error(),
						},
					}
					continue
				}
			}

			c.Lock()
			v, err := h.Func(c, in)
			c.Unlock()
			if err == ErrDisconnect {
				return
			}
			if err != nil {
				c.out <- &messageOut{
					ID: m.ID,
					Data: map[string]interface{}{
						"error": err.Error(),
					},
				}
				continue
			}

			h.RunAfter(c)
			c.out <- &messageOut{
				ID:   m.ID,
				Data: v,
			}
		}
	}()

	for m := range c.out {
		err := conn.WriteJSON(m)
		if err != nil {
			log.Println("conn.WriteJSON:", err)
			return
		}
	}
}
