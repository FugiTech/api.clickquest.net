package handlers

import (
	"errors"
	"time"

	"../api"
)

func init() {
	api.RegisterHandler("chat", chatParams{}, chat)
}

type chatParams struct {
	Message string `json:"message"`
}

func chat(c *api.Conn, i interface{}) (interface{}, error) {
	if c.User == nil {
		return nil, errors.New("Not logged in")
	}

	params := i.(*chatParams)
	cl := api.ChatLine{
		Name:     c.User.Name,
		Message:  params.Message,
		Color:    c.User.Color,
		Level:    api.LevelForClicks(c.User.Clicks, c.User.Hardcore > 0),
		Time:     time.Now(),
		Admin:    c.User.Admin,
		Mod:      c.User.Mod,
		Hardcore: c.User.Hardcore > 0,
	}

	_, err := api.DB.Exec("INSERT INTO chat(name,message,color,level,ip,time,admin,`mod`,hardcore) VALUES(?,?,?,?,?,?,?,?,?)",
		cl.Name, cl.Message, cl.Color, cl.Level, c.IP, cl.Time, cl.Admin, cl.Mod, cl.Hardcore)
	if err != nil {
		return nil, err
	}

	api.AppendChat(cl)
	api.Broadcast("chat", []api.ChatLine{cl})
	return nil, nil
}
