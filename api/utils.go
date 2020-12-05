package api

import (
	"database/sql"
	"math"
)

func ParseChat(rows *sql.Rows) ([]ChatLine, error) {
	var chat []ChatLine
	for rows.Next() {
		l := &ChatLine{}
		err := rows.Scan(&l.Name, &l.Message, &l.Color, &l.Level, &l.Time, &l.Admin, &l.Mod, &l.Hardcore)
		if err != nil {
			return nil, err
		}
		chat = append(chat, *l)
	}
	return chat, rows.Err()
}

func LevelForClicks(clicks uint64, hardcore bool) int {
	var initial, rate float64
	if hardcore {
		initial, rate = 200, 1.110316
	} else {
		initial, rate = 100, 1.0906595
	}
	return int(math.Floor(math.Log(1+(rate-1)*(float64(clicks)+0.5)/initial) / math.Log(rate)))
}
