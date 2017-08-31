package api

import (
	"database/sql"
	"log"
	"sync"
	"time"
)

var (
	DB          *sql.DB
	latestStats *Stats
	latestChat  []ChatLine
	Mutex       sync.Mutex
)

const LinesPerPage = 50

func Start() {
	Mutex.Lock()
	defer Mutex.Unlock()
	calculateStats()
	loadChat()

	go func() {
		c := time.Tick(1 * time.Minute)
		for range c {
			func() {
				Mutex.Lock()
				defer Mutex.Unlock()
				calculateStats()
				Broadcast("stats", latestStats)
			}()
		}
	}()
}

func getStats() *Stats {
	Mutex.Lock()
	defer Mutex.Unlock()
	return latestStats
}

func getChat() []ChatLine {
	Mutex.Lock()
	defer Mutex.Unlock()
	return latestChat
}

func AppendChat(c ChatLine) {
	Mutex.Lock()
	defer Mutex.Unlock()
	latestChat = append(latestChat, c)
	latestChat = latestChat[1:]
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
	latestChat, err = ParseChat(rows)
	if err != nil {
		log.Println("loadChat:", err)
		return
	}
	// reverse the slice since we went in DESC order
	for i, j := 0, len(latestChat)-1; i < j; i, j = i+1, j-1 {
		latestChat[i], latestChat[j] = latestChat[j], latestChat[i]
	}
}
