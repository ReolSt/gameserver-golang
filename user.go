package main

import (
  "database/sql"
  "log"
  _ "github.com/go-sql-driver/mysql"
)

type User struct {
  id string
  name string
  password string
}

func GetUser(db *sql.DB, id string) (User, error) {
  row := db.QueryRow("select * from user where id=?", id)
  var user User
  err := row.Scan(&user.id, &user.name, &user.password)

  return user, err
}

func CreateUser(
  db *sql.DB,
  user User) error {
    stmt, err := db.Prepare("INSERT INTO user(id, password, name) VALUES(?, ?, ?)")

    if err != nil {
      log.Fatal(err)
      return err
    }

    defer stmt.Close()

    _, err = stmt.Exec(user.id, user.name, user.password)
    return err
}
