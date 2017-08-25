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

func checkErr(err error) {
	if err != nil {
		log.Printf("Error %v ", err)
	}
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
func GetUserByKey(key string) string {
	db := getDbConn()
	rows, err := db.Query("SELECT login FROM access_keys JOIN users ON user_id=id WHERE key=$1 LIMIT 1;", key)
	checkErr(err)

	for rows.Next() {
		var username string
		err = rows.Scan(&username)
		checkErr(err)
		return username
	}

	return ""
}

// CreateUser adds user to database
func CreateUser(login string, password string) (int, error) {
	db := getDbConn()
	rows, err := db.Query("SELECT COUNT(*) FROM users WHERE login=$1 LIMIT 1;", login)
	checkErr(err)
	for rows.Next() {
		var userCount int
		err = rows.Scan(&userCount)
		checkErr(err)
		if userCount != 0 {
			return 0, errors.New("user already exists")
		}
	}

	hash, _ := hashPassword(password)

	var lastInserted int
	err = db.QueryRow("INSERT INTO users(login,password) VALUES($1,$2) returning id;", login, hash).Scan(&lastInserted)
	checkErr(err)

	return lastInserted, nil
}

// ValidateUser checks if user with login-password exists returning id or zero if doesn`t
func ValidateUser(login string, password string) int {
	db := getDbConn()
	rows, err := db.Query("SELECT id, password FROM users WHERE login=$1 LIMIT 1;", login)
	checkErr(err)

	for rows.Next() {
		var userID int
		var dbHash string
		err = rows.Scan(&userID, &dbHash)
		checkErr(err)
		if checkPasswordHash(password, dbHash) {
			return userID
		}
	}

	return 0
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
