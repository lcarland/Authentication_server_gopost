package db

import (
	"context"
	"fmt"
	"log"
	"os"
	_ "time"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq"

	_ "authapi/utils"
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
	var pguser string = os.Getenv("POSTGRES_USER")
	var pgpw string = os.Getenv("POSTGRES_PASSWORD")

	connString := fmt.Sprintf("postgres://%s:%s@%s:%d/%s", pguser, pgpw, host, port, dbname)

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

type country struct {
	Code  string
	Name  string
	Phone string
}

func (db *Db) GetCountry(code string) (*country, error) {
	query := "SELECT code, country, dialcode " +
		"FROM countries WHERE code = $1;"
	var c country
	err := db.QueryRow(context.Background(), query, code).Scan(
		&c.Code, &c.Name, &c.Phone,
	)
	if err != nil {
		fmt.Println("DB Error Occurred:", err)
		return nil, err
	}
	return &c, nil
}

// type NewUser struct {
// 	Username, Password, FirstName, LastName,
// 	Email, Phone, Country string
// 	IsSuper, IsStaff bool
// }

// func (db *Db) InsertUser(u NewUser) error {
// 	query := "INSERT INTO users " +
// 		"(username, passwordHash, first_name, last_name, email, " +
// 		"phone, country, is_superuser, is_staff) VALUES " +
// 		"($1, $2, $3, $4, $5, $6, $7, $8, $9);"
// 	pwHash := GetPasswordHash(u.Password)

// 	_, err := db.Exec(query,
// 		u.Username, pwHash, u.FirstName, u.LastName, u.Email,
// 		u.Phone, u.Country, u.IsSuper, u.IsStaff,
// 	)
// 	if err != nil {
// 		fmt.Println("DB INSERT Error:", err)
// 		return err
// 	}
// 	return nil
// }

// type User struct {
// 	Id int
// 	Username, FirstName, LastName,
// 	Email, Phone, Country string
// 	IsSuper, IsStaff, Is_active bool
// 	DateJoined, LastLogin       time.Time
// }

// func userQueryConstructor(cols string, selector string) string {
// 	return fmt.Sprintf("SELECT %s FROM users WHERE %s;", cols, selector)
// }

// var userPrivate string = "id, username, first_name, last_name, email, " +
// 	"phone, country, is_superuser, is_staff, is_active, date_joined, " +
// 	"last_login"

// var userPublic string = "id, username, email, country, is_active, date_joined"

// func (db *Db) SelectUserById(id int) (*User, error) {
// 	query := userQueryConstructor(userPrivate, "id = $1")
// 	var u User

// 	err := db.QueryRow(query, id).Scan(
// 		&u.Id, &u.Username, &u.FirstName, &u.LastName, &u.Email,
// 		&u.Phone, &u.Country, &u.IsSuper, &u.IsStaff, &u.Is_active,
// 		&u.DateJoined, &u.LastLogin,
// 	)
// 	if err != nil {
// 		fmt.Println("DB SELECT Error:", err)
// 		return nil, err
// 	}
// 	return &u, nil
// }

// type UserSession struct {
// 	SessionId string
// }

// func sessionQuery(col string) string {
// 	return fmt.Sprintf("SELECT session_id "+
// 		"FROM users WHERE %s;", col)
// }

// func (db *Db) SelectUserSession() {

// }
