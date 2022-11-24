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
    stmt, err := db.Prepare("insert into user(id, password, name) values(?, ?, ?)")

    if err != nil {
      log.Fatal(err)
      return err
    }

    defer stmt.Close()

    _, err = stmt.Exec(user.id, user.password, user.name)
    return err
}

func RemoveUser(
  db *sql.DB,
  id string) error {
    stmt, err := db.Prepare("delete from user where id=(?)")

    if err != nil {
      log.Fatal(err)
      return err
    }

    defer stmt.Close()

    _, err = stmt.Exec(id)
    return err
}
