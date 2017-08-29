package handlers

import (
	"errors"
	"fmt"
	"strings"

	"../api"
)

func init() {
	api.RegisterHandler("setcolor", setcolorParams{}, setcolor).After(api.AfterBroadcast)
}

type setcolorParams struct {
	Color string `json:"color"`
}

type setcolorResult struct {
	Color string
}

func setcolor(c *api.Conn, i interface{}) (interface{}, error) {
	if c.User == nil {
		return nil, errors.New("Not logged in")
	}

	params := i.(*setcolorParams)
	curColor, curType := api.GetColor(c.User.Color)
	selColor, selType := api.GetColor(params.Color)
	level := api.LevelForClicks(c.User.Clicks, c.User.Hardcore > 0)

	// Validation Checks
	switch {
	case selColor.Name == "default":
		return nil, errors.New("Invalid color")
	case level >= 75:
		// Pass, 75+ can set any color they want any time
	case curColor.Name == "default" && selType != "Normal":
		return nil, errors.New("Select a color before a shade")
	case level < 50 && selType != "Normal":
		return nil, errors.New("You're not allowed to choose a shade yet")
	case curColor.Name != "default" && curColor.Name != selColor.Name:
		return nil, errors.New("You can't change colors")
	case curType != "Normal":
		return nil, errors.New("You've already chosen your shade")
	}

	color := strings.ToUpper(params.Color)
	_, err := api.DB.Exec("UPDATE users SET color = ? WHERE id = ?", color, c.User.ID)
	if err != nil {
		return nil, fmt.Errorf("Failed to set color: %v", err)
	}
	c.User.Color = color

	return setcolorResult{color}, nil
}
