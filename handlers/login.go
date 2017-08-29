package handlers

import (
	_md5 "crypto/md5"
	_sha1 "crypto/sha1"
	"database/sql"
	"encoding/hex"
	"fmt"
	"regexp"
	"strconv"
	"time"

	"golang.org/x/crypto/bcrypt"

	"../api"
)

func init() {
	api.RegisterHandler("login", loginParams{}, login).After(api.AfterLogin)
}

type loginParams struct {
	Username       string `json:"username"`
	Password       string `json:"password"`
	IsRegistration bool   `json:"register"`
}

func login(c *api.Conn, i interface{}) (interface{}, error) {
	params := i.(*loginParams)

	user := &api.User{}
	var password string
	var oldpassword *string
	r := api.DB.QueryRow("SELECT `id`, `username`, `password`, `oldpassword`, `clicks`, `modified`, `color`, `admin`, `mod`, `banned`, `hardcore`, `totaltime` FROM users WHERE username = ?", params.Username)
	err := r.Scan(&user.ID, &user.Name, &password, &oldpassword, &user.ActualClicks, &user.ModifiedClicks, &user.Color, &user.Admin, &user.Mod, &user.Banned, &user.Hardcore, &user.TotalTime)
	if err == sql.ErrNoRows {
		if params.IsRegistration {
			return register(c, params)
		}
		return nil, fmt.Errorf("User %q does not exist", params.Username)
	}
	if err != nil {
		return nil, err
	}
	if params.IsRegistration {
		return nil, fmt.Errorf("User %q already exists", params.Username)
	}

	// In the past passwords were stored using a weak hash: md5(sha1(md5(sha1(md5(...)))))
	// so the first time the user logs in, we upgrade them to a stronger hash
	if oldpassword != nil {
		if *oldpassword != md5(sha1(md5(sha1(md5(params.Password))))) {
			return nil, fmt.Errorf("Invalid Password")
		}

		bytes, err := bcrypt.GenerateFromPassword([]byte(params.Password), 14)
		if err != nil {
			return nil, fmt.Errorf("Could not upgrade password security: %v", err)
		}

		password = string(bytes)
		_, err = api.DB.Exec("UPDATE users SET password = ?, oldpassword = NULL WHERE id = ?", password, user.ID)
		if err != nil {
			return nil, fmt.Errorf("Could not upgrade password security: %v", err)
		}
	}

	// Do the normal check (doing this right after upgrading is a waste but else blocks are icky)
	err = bcrypt.CompareHashAndPassword([]byte(password), []byte(params.Password))
	if err != nil {
		return nil, fmt.Errorf("Invalid Password")
	}

	user.Clicks = uint64(user.ActualClicks + int64(user.ModifiedClicks))
	user.SessionStart = time.Now()
	c.User = user

	return user, nil
}

var usernameRe = regexp.MustCompile(`^\w{0,16}$`)

func register(c *api.Conn, params *loginParams) (interface{}, error) {
	if !usernameRe.MatchString(params.Username) {
		return nil, fmt.Errorf("Username %q is invalid: must be 16 or less alphanumeric or underscore characters", params.Username)
	}
	password, err := bcrypt.GenerateFromPassword([]byte(params.Password), 14)
	if err != nil {
		return nil, err
	}

	r, err := api.DB.Exec("INSERT INTO users(username, password) VALUES(?,?)", params.Username, string(password))
	if err != nil {
		return nil, err
	}

	id, err := r.LastInsertId()
	if err != nil {
		return nil, err
	}

	user := &api.User{
		ID:           strconv.Itoa(int(id)),
		Name:         params.Username,
		Color:        "#FFFFFF",
		SessionStart: time.Now(),
	}
	c.User = user

	return user, nil
}

func md5(s string) string {
	h := _md5.Sum([]byte(s))
	return hex.EncodeToString(h[:])
}

func sha1(s string) string {
	h := _sha1.Sum([]byte(s))
	return hex.EncodeToString(h[:])
}
