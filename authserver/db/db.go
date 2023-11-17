package db

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"

	. "authapi/utils"
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

func (db *Db) GetCountry(code string) (*country, error) {
	query := "SELECT code, country, dialcode " +
		"FROM countries WHERE code = $1;"
	var c country
	err := db.QueryRow(query, code).Scan(
		&c.Code, &c.Name, &c.Phone,
	)
	if err != nil {
		fmt.Println("DB Error Occurred:", err)
		return nil, err
	}
	return &c, nil
}

type NewUser struct {
	Username, Password, FirstName, LastName,
	Email, Phone, Country string
	IsSuper, IsStaff bool
}

func (db *Db) InsertUser(u NewUser) error {
	query := "INSERT INTO users " +
		"(username, passwordHash, first_name, last_name, email, " +
		"phone, country, is_superuser, is_staff) VALUES " +
		"($1, $2, $3, $4, $5, $6, $7, $8, $9);"
	pwHash := GetPasswordHash(u.Password)

	_, err := db.Exec(query,
		u.Username, pwHash, u.FirstName, u.LastName, u.Email,
		u.Phone, u.Country, u.IsSuper, u.IsStaff,
	)
	if err != nil {
		fmt.Println("DB INSERT Error:", err)
		return err
	}
	return nil
}

type User struct {
	Id int
	Username, FirstName, LastName,
	Email, Phone, Country string
	IsSuper, IsStaff, Is_active bool
	DateJoined, LastLogin       time.Time
	SessionId                   string
}

func userQueryConstructor(selector string) string {
	return fmt.Sprintf("SELECT id, username, first_name, last_name, email, "+
		"phone, country, is_superuser, is_staff, is_active, date_joined, "+
		"last_login, session_id "+
		"FROM users WHERE %s;", selector)
}

func (db *Db) SelectUserById(id int) (*User, error) {
	query := userQueryConstructor("id = $1")
	var u User

	err := db.QueryRow(query, id).Scan(
		&u.Id, &u.Username, &u.FirstName, &u.LastName, &u.Email,
		&u.Phone, &u.Country, &u.IsSuper, &u.IsStaff, &u.DateJoined,
		&u.LastLogin, &u.SessionId,
	)
	if err != nil {
		fmt.Println("DB SELECT Error:", err)
		return nil, err
	}
	return &u, nil
}

func (db *Db) SelectUserByName(n string) (*User, error) {
	query := userQueryConstructor("username = $1")
	var u User

	err := db.QueryRow(query, n).Scan(
		&u.Id, &u.Username, &u.FirstName, &u.LastName, &u.Email,
		&u.Phone, &u.Country, &u.IsSuper, &u.IsStaff, &u.DateJoined,
		&u.LastLogin, &u.SessionId,
	)
	if err != nil {
		fmt.Println("DB SELECT Error:", err)
		return nil, err
	}
	return &u, nil
}

type UserSession struct {
	SessionId string
}

func sessionQuery(col string) string {
	return fmt.Sprintf("SELECT session_id "+
		"FROM users WHERE %s;", col)
}

func (db *Db) SelectUserSession() {

}
