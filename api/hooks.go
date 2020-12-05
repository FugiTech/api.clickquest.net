package api

func AfterLogin(c *Conn) {
	defer broadcastPlayers()

	connLock.RLock()
	defer connLock.RUnlock()
	c.Lock()
	defer c.Unlock()
	for _, conn := range connected {
		if c == conn {
			continue
		}
		func(c, conn *Conn) {
			conn.Lock()
			defer conn.Unlock()
			if c.User != nil && conn.User != nil && c.User.ID == conn.User.ID {
				conn.User = nil
				conn.out <- &messageOut{Type: "logout"}
			}
		}(c, conn)
	}

	c.LastHeartbeat = c.User.SessionStart
	c.LastActivity = c.User.SessionStart
	c.User.lastLevel = LevelForClicks(c.User.Clicks, c.User.Hardcore > 0)
}

func AfterBroadcast(c *Conn) {
	broadcastPlayers()
}

func AfterHeartbeat(c *Conn) {
	shouldBroadcast := func() bool {
		c.Lock()
		defer c.Unlock()
		if c.User != nil {
			level := LevelForClicks(c.User.Clicks, c.User.Hardcore > 0)
			shouldBroadcast := level != c.User.lastLevel
			c.User.lastLevel = level
			return shouldBroadcast
		}
		return false
	}()

	if shouldBroadcast {
		broadcastPlayers()
	}
}
