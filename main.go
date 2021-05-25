package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	"github.com/upper/db/v4"
	"github.com/upper/db/v4/adapter/mysql"
)

/*
type Genre struct {
	ID        uint   `db:"id"`
	GenreName string `db:"genre"`
}

type Book struct {
	ID      uint    `db:"id,omitempty"`
	Name    string  `db:"name"`
	Price   float64 `db:"author_id"`
	GenreID uint    `db:"subject_id"`
	Amount  uint
}
*/
var settings = mysql.ConnectionURL{
	User:     `root`,
	Password: `jelszavam`,
	Database: `bookstore`,
	Host:     `localhost`,
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func executeSqlScript(script string, sess db.Session) {
	commands := strings.Split(script, ";")
	for _, v := range commands {
		s := strings.TrimSpace(v)
		if 0 < len(s) {
			sess.SQL().Exec(s)
		}
	}
}

func main() {
	sql, err := ioutil.ReadFile("setup.sql")
	check(err)
	setupScript := string(sql)

	// Set logging level to DEBUG
	db.LC().SetLevel(db.LogLevelDebug)

	// Use Open to access the database.
	sess, err := mysql.Open(settings)
	if err != nil {
		log.Fatal("Open: ", err)
	}
	defer sess.Close()

	fmt.Println("Executing script:")
	executeSqlScript(setupScript, sess)
}
