package db

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq"

	"authapi/utils"
)

const (
	host   = "localhost"
	dbname = "authdb"
)

var port string = os.Getenv("PG_PORT")

type Db struct {
	*pgxpool.Pool
}

func DbService() *Db {
	var PGUSER string = os.Getenv("POSTGRES_USER")
	var PGPASSWD string = os.Getenv("POSTGRES_PASSWORD")

	connString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", PGUSER, PGPASSWD, host, port, dbname)

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
}

func (db *Db) InsertUser(u NewUser) error {
	query := "INSERT INTO users " +
		"(username, passwordHash, first_name, last_name, email, " +
		"phone, country) VALUES " +
		"($1, $2, $3, $4, $5, $6, $7);"
	pwHash := utils.GetPasswordHash(u.Password)

	_, err := db.Exec(context.Background(), query,
		u.Username, pwHash, u.FirstName, u.LastName, u.Email,
		u.Phone, u.Country,
	)
	if err != nil {
		fmt.Println("DB INSERT Error:", err)
		return err
	}
	return nil
}

//=====================================//
// ---- Query String Constructors ---- //
//=====================================//

func queryConstructor(table string, cols string, selector string) string {
	return fmt.Sprintf("SELECT %s FROM %s WHERE %s;", cols, table, selector)
}

func updateConstructor(table string, val string, selector string) string {
	return fmt.Sprintf("UPDATE %s SET %s WHERE %s", table, val, selector)
}

func deleteConstructor(table string, selector string) string {
	return fmt.Sprintf("DELETE FROM %s WHERE %s;", table, selector)
}

var userPrivate string = "id, username, first_name, last_name, email, " +
	"phone, country, is_superuser, is_staff, is_active, date_joined, " +
	"last_login"

var userPublic string = "id, username, country, is_active, date_joined"

//=================================//
// ---- User Table Management ---- //
//=================================//

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

// Get user private info. Protect for each user
func (db *Db) SelectPrivateUserById(id int) (*User, error) {
	query := queryConstructor("users", userPrivate, "id = $1")

	rows, _ := db.Query(context.Background(), query, id)
	u, err := pgx.CollectExactlyOneRow[User](rows, pgx.RowToStructByName[User])
	if err != nil {
		fmt.Println("DB SELECT Error:", err)
		return nil, err
	}
	return &u, nil
}

type UserPublic struct {
	Id         int       `db:"id"`
	Username   string    `db:"username"`
	Country    string    `db:"country"`
	DateJoined time.Time `db:"date_joined"`
	IsActive   bool      `db:"is_active"`
}

// Public Info about a user
func (db *Db) SelectPublicUser(id int) (*UserPublic, error) {
	query := queryConstructor("users", userPublic, "id = $1")
	rows, _ := db.Query(context.Background(), query, id)
	u, err := pgx.CollectExactlyOneRow[UserPublic](rows, pgx.RowToStructByName[UserPublic])
	if err == nil {
		fmt.Println("DB SEL Err", err)
		return &UserPublic{}, err
	}
	return &u, nil
}

type UserAuth struct {
	Id           int    `db:"id"`
	Username     string `db:"username"`
	PasswordHash string `db:"passwordHash"`
	IsSuperuser  bool   `db:"is_superuser"`
	IsStaff      bool   `db:"is_staff"`
	IsActive     bool   `db:"is_active"`
}

// Get user information prevelant to authentication and permissions
func (db *Db) SelectUserAuth(username string) (*UserAuth, error) {
	fields := "id, username, passwordHash, is_superuser, is_staff, is_active"
	query := queryConstructor("users", fields, "username = $1")
	rows, _ := db.Query(context.Background(), query, username)
	s, err := pgx.CollectExactlyOneRow[UserAuth](rows, pgx.RowToStructByName[UserAuth])
	if err != nil {
		return nil, err
	}
	println(s.PasswordHash)
	return &s, nil
}

type userId struct {
	Id int
}

func (db *Db) GetUserId(username string) int {
	query := queryConstructor("users", "id", "username = $1")
	rows, _ := db.Query(context.Background(), query, username)
	id, err := pgx.CollectExactlyOneRow[userId](rows, pgx.RowToStructByName[userId])
	if err != nil {
		fmt.Println(err)
		return 0
	}
	return id.Id
}

func (db *Db) GetUserIdWithEmail(email string) int {
	query := queryConstructor("users", "id", "email = $1")
	rows, _ := db.Query(context.Background(), query, email)
	id, err := pgx.CollectExactlyOneRow[userId](rows, pgx.RowToStructByName[userId])
	if err != nil {
		fmt.Println(err)
		return 0
	}
	return id.Id
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

// Updates user info with map. Validate map prior to calling func.
func (db *Db) UpdateUserProfile(id int, updates map[string]any) error {
	var updateSegment []string
	var argNumber int = 2
	var args []any
	args = append(args, id)

	for key, val := range updates {
		setVal := fmt.Sprintf("%s = $%d", key, argNumber)
		updateSegment = append(updateSegment, setVal)
		args = append(args, val)
		argNumber += 1
	}
	setSegment := strings.Join(updateSegment, ", ")
	query := updateConstructor("users", setSegment, "id = $1")
	_, err := db.Exec(context.Background(), query, args...)
	if err != nil {
		return err
	}
	return nil
}

func (db *Db) NewUserHashById(id int, password string) error {
	hash := utils.GetPasswordHash(password)
	query := updateConstructor("users", "passwordHash = $2", "id = $1")
	_, err := db.Exec(context.Background(), query, id, hash)
	if err != nil {
		return err
	}
	return nil
}

func (db *Db) UpdateUserLoginTime(id int) error {
	query := updateConstructor("users", "last_login = $2", "id = $1")
	timeStamp := time.Now().UTC().Format(time.RFC3339)
	_, err := db.Exec(context.Background(), query, id, timeStamp)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func (db *Db) DeleteUser(id int) error {
	query := deleteConstructor("users", "id = $1")
	_, err := db.Exec(context.Background(), query, id)
	if err != nil {
		return err
	}
	return nil
}

//====================================//
// ---- Session table management ---- //
//====================================//

func (db *Db) NewUserSession(id int, token string, pwReset bool) error {
	query := "INSERT INTO sessions (token, user_id, pw_reset, expires) VALUES ($1, $2, $3, $4)"
	var expire time.Time
	if pwReset {
		expire = time.Now().UTC().Add(time.Minute * 5)
	} else {
		expire = time.Now().UTC().Add(time.Hour * 720) // 30 days
	}
	_, err := db.Exec(context.Background(), query, token, id, pwReset, expire)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

type sessionCheck struct {
	Valid   bool      `db:"valid"`
	User_id int       `db:"user_id"`
	PwReset bool      `db:"pw_reset"`
	Expires time.Time `db:"expires"`
}

// session check with QueryToken func returns a bool and error.
// pwReset is false for refresh tokens and true for password reset tokens
// look for 3 possible outcomes:
//  1. true, no error
//     - user is authorized, invalidate session and refresh jwt
//  2. false, no error
//     - session is assumed hijacked, delete all user tokens
//  3. false, error( 'ErrNoRows' )
//     - token was removed, user is asked to login again
func (db *Db) QueryToken(token string, id int, pwReset bool) (bool, error) {
	query := queryConstructor("sessions", "valid, user_id, pw_reset, expires", "token = $1")
	rows, _ := db.Query(context.Background(), query, token)
	s, err := pgx.CollectExactlyOneRow[sessionCheck](rows, pgx.RowToStructByName[sessionCheck])
	if err != nil || s.User_id != id {
		fmt.Println(err)
		return false, err
	}

	if time.Now().UTC().After(s.Expires) {
		return false, nil
	}

	if pwReset && s.PwReset {
		return true, nil
	} else if pwReset && !s.PwReset {
		return false, nil
	}

	if !s.Valid {
		return false, nil
	}

	return true, nil
}

// invalidate the old refresh token after it has been used
func (db *Db) InvalidateSession(token string) error {
	query := updateConstructor("sessions", "valid = FALSE", "token = $1")
	_, err := db.Exec(context.Background(), query, token)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

// Invalidate all user sessions based on id.
// Used to force a user to login.
// This should be triggered after an invalidated token has been reused.
func (db *Db) InvalidateAllSessions(id int) error {
	query := deleteConstructor("sessions", "user_id = $1")
	_, err := db.Exec(context.Background(), query, id)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

// delete single session. Used for logging out
func (db *Db) DeleteSession(token string) error {
	query := deleteConstructor("sessions", "token = $1")
	_, err := db.Exec(context.Background(), query, token)
	if err != nil {
		return err
	}
	return nil
}

// Only for testing purposes. Need to disable for production use, dynamically or manually
func (db *Db) DeleteAllUsers() error {
	query := "DELETE FROM users;"
	_, err := db.Exec(context.Background(), query)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}
