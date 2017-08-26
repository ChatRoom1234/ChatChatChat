package db

import (
	"database/sql"
	"errors"
	"fmt"
	"io/ioutil"
	"log"

	yaml "gopkg.in/yaml.v2"

	// vasilesk: don`t understand why we need it
	//           but nothing works without this import
	_ "github.com/lib/pq"
)

var db *sql.DB

type conf struct {
	User     string `yaml:"user"`
	Database string `yaml:"database"`
	Password string `yaml:"password"`
}

func logErr(err error) {
	log.Printf("Error %v ", err)
}

func getConf() conf {

	yamlFile, err := ioutil.ReadFile("config/db.yaml")
	if err != nil {
		log.Printf("yamlFile.Get err   #%v ", err)
	}

	var c conf
	err = yaml.Unmarshal(yamlFile, &c)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}

	return c
}

// CreateKey creates key for user by his id
func CreateKey(userID int) string {
	key := randKey()

	db := getDbConn()
	db.QueryRow("INSERT INTO access_keys(user_id,key) VALUES($1,$2);", userID, key)

	return key
}

// GetUserByKey returns user login by key or "" if no user with the key exists
func GetUserByKey(key string) (int, string) {
	db := getDbConn()
	rows, err := db.Query("SELECT user_id, login FROM access_keys JOIN users ON user_id=id WHERE key=$1 LIMIT 1;", key)
	if err != nil {
		logErr(err)
		return 0, ""
	}

	var userID int
	var username string
	for rows.Next() {
		err = rows.Scan(&userID, &username)
		if err != nil {
			logErr(err)
			return 0, ""
		}
		return userID, username
	}
	return 0, ""
}

// CreateUser adds user to database
func CreateUser(login string, password string) (int, error) {
	db := getDbConn()
	rows, err := db.Query("SELECT COUNT(*) FROM users WHERE login=$1 LIMIT 1;", login)
	if err != nil {
		logErr(err)
		return 0, errors.New("DB error")
	}
	for rows.Next() {
		var userCount int
		err = rows.Scan(&userCount)
		if err != nil {
			logErr(err)
			return 0, errors.New("Backend error")
		}
		if userCount != 0 {
			return 0, errors.New("user already exists")
		}
	}

	hash, _ := hashPassword(password)

	var lastInserted int
	err = db.QueryRow("INSERT INTO users(login,password) VALUES($1,$2) returning id;", login, hash).Scan(&lastInserted)
	if err != nil {
		logErr(err)
		return 0, errors.New("DB error")
	}

	return lastInserted, nil
}

// ValidateUser checks if user with login-password exists returning id or zero if doesn`t
func ValidateUser(login string, password string) int {
	db := getDbConn()
	rows, err := db.Query("SELECT id, password FROM users WHERE login=$1 LIMIT 1;", login)
	if err != nil {
		logErr(err)
		return 0
	}

	for rows.Next() {
		var userID int
		var dbHash string
		err = rows.Scan(&userID, &dbHash)
		if err != nil {
			logErr(err)
			return 0
		}
		if checkPasswordHash(password, dbHash) {
			return userID
		}
	}

	return 0
}

// GetHistory returns message history
func GetHistory() ([100]string, error) {
	var result [100]string
	db := getDbConn()
	rows, err := db.Query("SELECT login, message FROM history inner join users on users.id = user_id ORDER BY history.id LIMIT 100;")
	if err != nil {
		logErr(err)
		return result, errors.New("DB error")
	}

	i := 0
	var username string
	var message string
	for rows.Next() {
		err = rows.Scan(&username, &message)
		if err != nil {
			logErr(err)
			return result, errors.New("Backend error")
		}
		result[i] = username + ": " + message
		i++
	}

	return result, nil
}

// AddMessage stores message into DB
func AddMessage(userID int, message []byte) error {
	db := getDbConn()
	db.QueryRow("INSERT INTO history(user_id, message) VALUES($1, $2);", userID, message)

	return nil
}

func initDb() error {
	c := getConf()
	dbinfo := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable port=5432 host=vasilesk.ru",
		c.User, c.Password, c.Database)
	var err error
	db, err = sql.Open("postgres", dbinfo)

	return err
}

func getDbConn() *sql.DB {
	fmt.Println("separate db conn")
	c := getConf()
	dbinfo := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable port=5432 host=vasilesk.ru",
		c.User, c.Password, c.Database)
	conn, _ := sql.Open("postgres", dbinfo)

	return conn
}

func init() {
	err := initDb()
	if err != nil {
		panic(err)
	}
	initRandKey()
}
