package api

import "time"

type User struct {
	Name         string    `json:"name"`
	Hardcore     int       `json:"hardcore"`
	Clicks       uint64    `json:"clicks"`
	Color        string    `json:"color"`
	SessionStart time.Time `json:"sessionStart"`
	TotalTime    uint64    `json:"totalTime"`

	ID             string `json:"-"`
	ActualClicks   int64  `json:"-"`
	ModifiedClicks uint64 `json:"-"`
	Admin          bool   `json:"-"`
	Mod            bool   `json:"-"`
	Banned         bool   `json:"-"`

	lastLevel int // Used for AfterHeartbeat hook to determine whether to broadcastPlayers
}

type Color struct {
	Name   string
	Normal string
	Dark   string
	Light  string
}

type Stats struct {
	Users         int          `json:"users"`
	TotalClicks   int          `json:"clicks"`
	AverageClicks int          `json:"avgClicks"`
	AverageLevel  int          `json:"avgLevel"`
	TopTen        []UserStats  `json:"topTen"`
	Colors        []ColorStats `json:"colors"`
	HallOfFame    []UserStats  `json:"hallOfFame"`
}
type UserStats struct {
	Name   string `json:"name"`
	Color  string `json:"color"`
	Clicks int    `json:"clicks"`
	Level  int    `json:"level"`
}
type ColorStats struct {
	Name          string `json:"name"`
	Color         string `json:"color"`
	Players       int    `json:"players"`
	MaxClicks     int    `json:"maxClicks"`
	TotalClicks   int    `json:"clicks"`
	AverageClicks int    `json:"avgClicks"`
}

type ChatLine struct {
	Name     string    `json:"name"`
	Message  string    `json:"message"`
	Color    string    `json:"color"`
	Level    int       `json:"level"`
	Time     time.Time `json:"time"`
	Admin    bool      `json:"admin"`
	Mod      bool      `json:"mod"`
	Hardcore bool      `json:"hardcore"`
}
