package db

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"
	_ "time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq"

	"authapi/utils"
)

const (
	host   = "localhost"
	port   = 5432
	dbname = "authdb"
)

type Db struct {
	*pgxpool.Pool
}

func DbService() *Db {
	var PGUSER string = os.Getenv("POSTGRES_USER")
	var PGPASSWD string = os.Getenv("POSTGRES_PASSWORD")

	connString := fmt.Sprintf("postgres://%s:%s@%s:%d/%s", PGUSER, PGPASSWD, host, port, dbname)

	dbpool, err := pgxpool.New(context.Background(), connString)
	if err != nil {
		log.Fatal(err)
	}

	err = dbpool.Ping(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	return &Db{dbpool}
}

type Country struct {
	Code  string `db:"code"`
	Name  string `db:"country"`
	Phone string `db:"dialcode"`
}

func (db *Db) GetCountry(code string) (*Country, error) {
	query := "SELECT code, country, dialcode " +
		"FROM countries WHERE code = $1;"
	rows, err := db.Query(context.Background(), query, code)
	if err != nil {
		fmt.Println("DB Error Occurred:", err)
		return nil, err
	}
	c, err := pgx.CollectExactlyOneRow[Country](rows, pgx.RowToStructByName[Country])
	if err != nil {
		fmt.Println("DB Error Occurred:", err)
		return nil, err
	}
	return &c, nil
}

func (db *Db) GetAllCountries() (*[]Country, error) {
	query := "SELECT code, country, dialcode FROM countries;"
	rows, _ := db.Query(context.Background(), query)
	c, err := pgx.CollectRows[Country](rows, pgx.RowToStructByName[Country])
	return &c, err
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
	pwHash := utils.GetPasswordHash(u.Password)

	_, err := db.Exec(context.Background(), query,
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
	Id         int       `db:"id"`
	Username   string    `db:"username"`
	FirstName  string    `db:"first_name"`
	LastName   string    `db:"last_name"`
	Email      string    `db:"email"`
	Phone      string    `db:"phone"`
	Country    string    `db:"country"`
	IsSuper    bool      `db:"is_superuser"`
	IsStaff    bool      `db:"is_staff"`
	Is_active  bool      `db:"is_active"`
	DateJoined time.Time `db:"date_joined"`
	LastLogin  time.Time `db:"last_login"`
}

func queryConstructor(table string, cols string, selector string) string {
	return fmt.Sprintf("SELECT %s FROM %s WHERE %s;", cols, table, selector)
}

func updateConstructor(table string, val string, selector string) string {
	return fmt.Sprintf("UPDATE %s SET %s WHERE %s", table, val, selector)
}

var userPrivate string = "id, username, first_name, last_name, email, " +
	"phone, country, is_superuser, is_staff, is_active, date_joined, " +
	"last_login"

var userPublic string = "id, username, email, country, is_active, date_joined"

func (db *Db) SelectUserById(id int) (*User, error) {
	query := queryConstructor("users", userPrivate, "id = $1")

	rows, _ := db.Query(context.Background(), query, id)
	u, err := pgx.CollectExactlyOneRow[User](rows, pgx.RowToStructByName[User])
	if err != nil {
		fmt.Println("DB SELECT Error:", err)
		return nil, err
	}
	return &u, nil
}

type userSession struct {
	SessionId string `db:"session_id"`
}

func (db *Db) SelectUserSession(id int) string {
	query := queryConstructor("users", "session_id", "id = $1")
	rows, _ := db.Query(context.Background(), query, id)
	s, err := pgx.CollectExactlyOneRow[userSession](rows, pgx.RowToStructByName[userSession])
	if err != nil {
		fmt.Println(err)
		return ""
	}
	return s.SessionId
}

func (db *Db) NewUserSession(id int) error {
	session, _ := utils.GenerateCryptoString()
	val := fmt.Sprintf("session_id = %s", session)
	query := updateConstructor("users", val, "id = $1")
	_, err := db.Exec(context.Background(), query, id)
	if err != nil {
		fmt.Println("Session Update Error,", err)
		return err
	}
	return nil
}

type userHash struct {
	PasswordHash string `db:"passwordHash"`
}

func (db *Db) SelectUserHash(id int) string {
	query := queryConstructor("users", "passwordHash", "id = $1")
	rows, _ := db.Query(context.Background(), query, id)
	h, err := pgx.CollectExactlyOneRow[userHash](rows, pgx.RowToStructByName[userHash])
	if err != nil {
		fmt.Println(err)
		return ""
	}
	return h.PasswordHash
}

func (db *Db) NewUserHash(id int, password string) error {
	hash := utils.GetPasswordHash(password)
	val := fmt.Sprintf("passwordHash = %s", hash)
	query := updateConstructor("users", val, "id = $1")
	_, err := db.Exec(context.Background(), query, id)
	if err != nil {
		return err
	}
	return nil
}
