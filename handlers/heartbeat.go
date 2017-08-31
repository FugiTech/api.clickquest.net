package handlers

import (
	"math"
	"time"

	"../api"
)

func init() {
	api.RegisterHandler("heartbeat", heartbeatParams{}, heartbeat).After(api.AfterHeartbeat)
}

type heartbeatParams struct {
	Clicks uint64 `json:"clicks"`
}

type heartbeatResult struct {
	ClickAdjustment int64
	Hardcore        int
	Color           string
}

func heartbeat(c *api.Conn, i interface{}) (interface{}, error) {
	if c.User == nil {
		return nil, nil
	}

	params := i.(*heartbeatParams)
	resp := &heartbeatResult{}

	now := time.Now()
	elapsed := now.Sub(c.LastHeartbeat)
	c.LastHeartbeat = now

	// Clicking too fast - the client is supposed to prevent this so this is likely illegitimate
	if params.Clicks-c.User.Clicks > uint64(math.Ceil(12*elapsed.Seconds())) {
		return nil, api.ErrDisconnect
	}

	if params.Clicks == c.User.Clicks {
		if now.Sub(c.LastActivity) > 2*time.Minute {
			return nil, api.ErrDisconnect
		}
		return nil, nil
	}

	c.LastActivity = now
	c.User.Clicks = params.Clicks

	// Sync with database changes
	oldModified := c.User.ModifiedClicks
	err := api.DB.QueryRow("SELECT modified, color, admin, `mod`, banned FROM users WHERE id = ?", c.User.ID).Scan(&c.User.ModifiedClicks, &c.User.Color, &c.User.Admin, &c.User.Mod, &c.User.Banned)
	if err != nil {
		return nil, err
	}
	if c.User.Banned {
		return nil, api.ErrDisconnect
	}
	resp.ClickAdjustment += int64(c.User.ModifiedClicks) - int64(oldModified)

	// Check if they should move on to hardcore mode
	if c.User.Hardcore == 0 && int64(c.User.Clicks)+resp.ClickAdjustment >= 6666666 {
		err = api.DB.QueryRow("SELECT max(hardcore)+1 FROM users").Scan(&c.User.Hardcore)
		if err != nil {
			return nil, err
		}
		resp.ClickAdjustment += -6666666
		resp.Hardcore = c.User.Hardcore
	}

	// Other checks & adjustments should go here

	// Adjust clicks based on prior checks
	c.User.Clicks = uint64(int64(c.User.Clicks) + resp.ClickAdjustment)
	c.User.ActualClicks = int64(c.User.Clicks) - int64(c.User.ModifiedClicks)
	level := api.LevelForClicks(c.User.Clicks, c.User.Hardcore > 0)

	_, err = api.DB.Exec("UPDATE users SET clicks = ?, level = ?, ip = ?, last_online = ?, totaltime = totaltime + ? WHERE id = ?", c.User.ActualClicks, level, c.IP, now, elapsed.Seconds(), c.User.ID)
	if err != nil {
		return nil, err
	}

	resp.Color = c.User.Color
	return resp, nil
}
