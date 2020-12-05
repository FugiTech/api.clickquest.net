package handlers

import (
	"github.com/fugiman/api.clickquest.net/api"
)

func init() {
	api.RegisterHandler("chatlog", chatlogParams{}, chatlog)
}

type chatlogParams struct {
	Page int `json:"page"`
}

type chatlogResult struct {
	Chat  []api.ChatLine `json:"chat"`
	Pages int            `json:"pages"`
}

func chatlog(c *api.Conn, i interface{}) (interface{}, error) {
	params := i.(*chatlogParams)
	r := &chatlogResult{}

	rows, err := api.DB.Query("SELECT name, message, color, level, time, admin, `mod`, hardcore FROM chat WHERE id > ? ORDER BY id ASC LIMIT ?", (params.Page-1)*api.LinesPerPage, api.LinesPerPage)
	if err != nil {
		return nil, err
	}
	r.Chat, err = api.ParseChat(rows)
	if err != nil {
		return nil, err
	}
	err = api.DB.QueryRow("SELECT CEILING(max(id) / ?) FROM chat", api.LinesPerPage).Scan(&r.Pages)
	return r, err
}
