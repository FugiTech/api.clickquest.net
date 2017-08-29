package api

func AfterLogin(c *Conn) {
	defer broadcastPlayers()

	connLock.RLock()
	defer connLock.RUnlock()
	c.Lock()
	for _, conn := range connected {
		if c == conn {
			continue
		}
		conn.Lock()
		if c.User != nil && conn.User != nil && c.User.ID == conn.User.ID {
			conn.User = nil
			conn.out <- &messageOut{Type: "logout"}
		}
		conn.Unlock()
	}

	c.LastHeartbeat = c.User.SessionStart
	c.LastActivity = c.User.SessionStart
	c.User.lastLevel = LevelForClicks(c.User.Clicks, c.User.Hardcore > 0)
	c.Unlock()
}

func AfterBroadcast(c *Conn) {
	broadcastPlayers()
}

func AfterHeartbeat(c *Conn) {
	c.Lock()
	var shouldBroadcast bool
	if c.User != nil {
		level := LevelForClicks(c.User.Clicks, c.User.Hardcore > 0)
		shouldBroadcast = level != c.User.lastLevel
		c.User.lastLevel = level
	}
	c.Unlock()

	if shouldBroadcast {
		broadcastPlayers()
	}
}
