package handlers

import "github.com/fugiman/api.clickquest.net/api"

func init() {
	api.RegisterHandler("logout", nil, logout).After(api.AfterBroadcast)
}

func logout(c *api.Conn, i interface{}) (interface{}, error) {
	c.User = nil
	return nil, nil
}
