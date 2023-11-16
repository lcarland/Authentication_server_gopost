package db

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "pguser"
	password = "b~pAa~pzZw&r-TjdN&N7&h7xF$CK"
	dbname   = "authdb"
)

type Db struct{ *sql.DB }

func DbService() *Db {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	dbConn, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal(err)
	}

	err = dbConn.Ping()
	if err != nil {
		log.Fatal(err)
	}
	return &Db{dbConn}
}

type country struct {
	Code  string
	Name  string
	Phone string
}

func (db *Db) GetCountry(code string) *country {
	query := "SELECT code, country, dialcode " +
		"FROM countries WHERE code = $1;"
	fmt.Println(query)
	var c country
	err := db.QueryRow(query, code).Scan(
		&c.Code, &c.Name, &c.Phone,
	)
	if err != nil {
		fmt.Println("DB Error Occurred:", err)
		return nil
	}
	return &c
}
