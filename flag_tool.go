// +build ignore

package main

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"

	_ "github.com/mattn/go-sqlite3"
)

const flagToolDoc = `ITU-Minitwit Tweet Flagging Tool

Usage:
  flag_tool <tweet_id>...
  flag_tool -i
  flag_tool -h
Options:
  -h            Show this screen.
  -i            Dump all tweets and authors to STDOUT.`

func main() {
	if len(os.Args) < 2 {
		fmt.Println(flagToolDoc)
		return
	}

	db, err := sql.Open("sqlite3", "/tmp/minitwit.db")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Can't open database: %s\n", err)
		os.Exit(1)
	}
	defer db.Close()

	switch os.Args[1] {
	case "-h":
		fmt.Println(flagToolDoc)
	case "-i":
		rows, err := db.Query("SELECT message_id, author_id, text, pub_date, flagged FROM message")
		if err != nil {
			fmt.Fprintf(os.Stderr, "SQL error: %s\n", err)
			os.Exit(1)
		}
		defer rows.Close()
		for rows.Next() {
			var msgID, authorID, flagged int
			var text string
			var pubDate int64
			rows.Scan(&msgID, &authorID, &text, &pubDate, &flagged)
			fmt.Printf("%d,%d,%s,%d\n", msgID, authorID, text, flagged)
		}
	default:
		for _, arg := range os.Args[1:] {
			id, err := strconv.Atoi(arg)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Invalid tweet ID: %s\n", arg)
				continue
			}
			_, err = db.Exec("UPDATE message SET flagged=1 WHERE message_id=?", id)
			if err != nil {
				fmt.Fprintf(os.Stderr, "SQL error: %s\n", err)
			} else {
				fmt.Printf("Flagged entry: %d\n", id)
			}
		}
	}
}
