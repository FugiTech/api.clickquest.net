package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"

	_ "github.com/go-sql-driver/mysql"
)

var (
	fixers = []*regexp.Regexp{
		regexp.MustCompile(`(\A|\s)(\*\S(?:.*?\S)?\*)(\s|\z)`),                 // Bold
		regexp.MustCompile(`(\A|\s)(_\S(?:.*?\S)?_)(\s|\z)`),                   // Italic
		regexp.MustCompile(`(\A|\s)(\[\S(?:.*?\S)?\]\(\S(?:.*?\S)?\))(\s|\z)`), // Links
	}
	spaceRe = regexp.MustCompile(`[\s\x00\x85\x{2424}\x{2028}]+`) // From https://github.com/markdown-it/markdown-it/blob/master/lib/rules_core/normalize.js
)

func main() {
	oldDB, err := sql.Open("mysql", "xxx/clickquest")
	if err != nil {
		log.Fatal("sql.Open(oldDB):", err)
	}
	newDB, err := sql.Open("mysql", "xxx/clickquest.net")
	if err != nil {
		log.Fatal("sql.Open(newDB):", err)
	}
	_ = newDB

	// Support resuming
	var maxID int
	newDB.QueryRow("SELECT max(id) FROM chat").Scan(&maxID)

	rows, err := oldDB.Query("SELECT id, name, message, color, level, ip, time FROM chat WHERE id > ? ORDER BY id ASC", maxID)
	if err != nil {
		log.Fatal("oldDB.Query:", err)
	}

	values := []interface{}{}
	for rows.Next() {
		var (
			id         int
			name       string
			message    string
			color      string
			level      int
			ip         string
			created_at int64

			admin    bool
			mod      bool
			hardcore bool
		)
		err = rows.Scan(&id, &name, &message, &color, &level, &ip, &created_at)
		if err != nil {
			log.Fatal("rows.Scan():", err)
		}

		name = strings.TrimSpace(name)
		nname := name
		hardcore = strings.HasPrefix(nname, "[H]")
		nname = strings.TrimPrefix(nname, "[H]")
		admin = strings.HasPrefix(nname, "&lt;ADMIN&gt;")
		nname = strings.TrimPrefix(nname, "&lt;ADMIN&gt;")
		mod = strings.HasPrefix(nname, "&lt;GM&gt;")
		nname = strings.TrimPrefix(nname, "&lt;GM&gt;")
		admin = admin || strings.HasPrefix(nname, "&lt;LRR&gt;") // Promote LRR to staff
		nname = strings.TrimPrefix(nname, "&lt;LRR&gt;")

		message = spaceRe.ReplaceAllString(message, " ")
		message = strings.TrimSpace(message)
		nmessage := message
		for _, re := range fixers {
			for re.MatchString(nmessage) {
				nmessage = re.ReplaceAllString(nmessage, `${1}\${2}${3}`)
			}
		}

		d, err := html.Parse(strings.NewReader(nmessage))
		if err != nil {
			log.Fatal("html.Parse():", err)
		}

		b := &bytes.Buffer{}
		var f func(*html.Node)
		f = func(n *html.Node) {
			if n.Type == html.DocumentNode {
				for c := n.FirstChild; c != nil; c = c.NextSibling {
					f(c)
				}
				return
			}
			if n.Type == html.TextNode {
				fmt.Fprint(b, n.Data)
				if n.FirstChild != nil {
					log.Fatal("Unexpected FirstChild for TextNode:", n, n.FirstChild)
				}
				return
			}
			if n.Type != html.ElementNode {
				log.Fatal("Unexpected node type:", n)
			}

			attrs := map[string]string{}
			for _, a := range n.Attr {
				attrs[a.Key] = a.Val
			}

			if _, ok := attrs["style"]; ok {
				fmt.Fprint(b, "[")
			}

			switch n.DataAtom {
			case atom.A:
				fmt.Fprint(b, "[")
			case atom.Img:
				fmt.Fprintf(b, "[%s][%s]", attrs["alt"], strings.Replace(attrs["src"], "]", "%5D", -1))
			case atom.I:
				fmt.Fprint(b, " _")
			case atom.B:
				fmt.Fprint(b, " *")
			case atom.U:
				fmt.Fprint(b, "[")
			case atom.Q:
				fmt.Fprint(b, "\"")
			case atom.Br, atom.Code, atom.Span:
				// Ignore it
			case atom.Html, atom.Head, atom.Body:
				// The parser injects these, annoyingly
			case atom.Script:
				fmt.Fprint(b, "<script>")
			case atom.Style:
				fmt.Fprint(b, "<style>")
			case atom.Div:
				fmt.Fprint(b, "<div")
				for _, a := range n.Attr {
					fmt.Fprintf(b, " %s=\"%s\"", a.Key, a.Val)
				}
				fmt.Fprint(b, ">")
			default:
				fmt.Fprintf(b, "<%s>", strings.ToUpper(n.Data))
			}

			for c := n.FirstChild; c != nil; c = c.NextSibling {
				f(c)
			}

			switch n.DataAtom {
			case atom.A:
				fmt.Fprintf(b, "](%s)", strings.Replace(attrs["href"], ")", "%29", -1))
			case atom.I:
				fmt.Fprint(b, "_ ")
			case atom.B:
				fmt.Fprint(b, "* ")
			case atom.U:
				fmt.Fprint(b, "]{text-decoration: underline}")
			case atom.Q:
				fmt.Fprint(b, "\"")
			case atom.Script:
				fmt.Fprint(b, "</script>")
			case atom.Style:
				fmt.Fprint(b, "</style>")
			case atom.Div:
				fmt.Fprint(b, "</div>")
			}

			if style, ok := attrs["style"]; ok {
				fmt.Fprintf(b, "]{%s}", style)
			}
		}
		f(d)
		nmessage = b.String()
		nmessage = spaceRe.ReplaceAllString(nmessage, " ")
		nmessage = strings.TrimSpace(nmessage)

		values = append(values, id, nname, nmessage, "#"+color, level, ip, time.Unix(created_at, 0), admin, mod, hardcore)
	}
	err = rows.Err()
	if err != nil {
		log.Fatal("rows.Err():", err)
	}

	for i := 0; i < len(values); i += 1000 {
		e := i + 1000
		if e > len(values) {
			e = len(values)
		}
		v := values[i:e]
		r := (e - i) / 10
		q := "INSERT INTO chat(id, name, message, color, level, ip, time, admin, `mod`, hardcore) VALUES"
		q += strings.Repeat(",(?,?,?,?,?,?,?,?,?,?)", r)[1:]
		_, err = newDB.Exec(q, v...)
		if err != nil {
			log.Fatal("newDB.Exec():", err)
		}
	}
}
