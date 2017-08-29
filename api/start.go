package api

import (
	"log"
	"time"
)

func Start() {
	Mutex.Lock()
	calculateStats()
	loadChat()
	Mutex.Unlock()

	go func() {
		c := time.Tick(1 * time.Minute)
		for range c {
			Mutex.Lock()
			calculateStats()
			Broadcast("stats", latestStats)
			Mutex.Unlock()
		}
	}()
}

func calculateStats() {
	stats := &Stats{}

	err := DB.QueryRow("SELECT count(id), sum(clicks+modified)+(sum(hardcore)*6666666), sum(level) FROM users WHERE banned=0").Scan(&stats.Users, &stats.TotalClicks, &stats.AverageLevel)
	if err != nil {
		log.Println("calculateStats:", err)
		return
	}
	stats.AverageClicks = stats.TotalClicks / stats.Users
	stats.AverageLevel = stats.AverageLevel / stats.Users // We stored total into average to make it easier

	rows, err := DB.Query("SELECT username, color, (clicks+modified), level FROM users WHERE banned = 0 AND hardcore=0 ORDER BY (clicks+modified) DESC LIMIT 10")
	if err != nil {
		log.Println("calculateStats:", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		u := &UserStats{}
		err = rows.Scan(&u.Name, &u.Color, &u.Clicks, &u.Level)
		if err != nil {
			log.Println("calculateStats:", err)
			return
		}
		stats.TopTen = append(stats.TopTen, *u)
	}
	err = rows.Err()
	if err != nil {
		log.Println("calculateStats:", err)
		return
	}

	for _, color := range colors {
		c := &ColorStats{
			Name:  color.Name,
			Color: color.Normal,
		}
		err = DB.QueryRow("SELECT count(id), max(clicks+modified), sum(clicks+modified) FROM users WHERE banned=0 AND hardcore=0 AND color IN (?, ?, ?)", color.Normal, color.Light, color.Dark).Scan(&c.Players, &c.MaxClicks, &c.TotalClicks)
		if err != nil {
			log.Println("calculateStats:", err)
			return
		}
		c.AverageClicks = c.TotalClicks / c.Players
		stats.Colors = append(stats.Colors, *c)
	}

	rows, err = DB.Query("SELECT username, color, (clicks+modified), level FROM users WHERE banned = 0 AND hardcore>0 ORDER BY hardcore ASC")
	if err != nil {
		log.Println("calculateStats:", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		u := &UserStats{}
		err = rows.Scan(&u.Name, &u.Color, &u.Clicks, &u.Level)
		if err != nil {
			log.Println("calculateStats:", err)
			return
		}
		stats.HallOfFame = append(stats.HallOfFame, *u)
	}
	err = rows.Err()
	if err != nil {
		log.Println("calculateStats:", err)
		return
	}

	latestStats = stats
}

func loadChat() {
	rows, err := DB.Query("SELECT name, message, color, level, time, admin, `mod`, hardcore FROM chat ORDER BY id DESC LIMIT 100")
	if err != nil {
		log.Println("loadChat:", err)
		return
	}
	defer rows.Close()
	LatestChat, err = ParseChat(rows)
	if err != nil {
		log.Println("loadChat:", err)
		return
	}
	// reverse the slice since we went in DESC order
	for i, j := 0, len(LatestChat)-1; i < j; i, j = i+1, j-1 {
		LatestChat[i], LatestChat[j] = LatestChat[j], LatestChat[i]
	}
}
