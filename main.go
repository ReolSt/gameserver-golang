package main

import (
  "database/sql"
  "net/http"
  "fmt"
  "time"

  _ "github.com/go-sql-driver/mysql"

  "context"
  "github.com/go-redis/redis/v9"

  "github.com/google/uuid"
)

var loginDB *sql.DB
var sessionDB *redis.Client
var ctx context.Context

type MySQLConfig struct {
  host string
  port uint16
  user string
  password string
}

type RedisConfig struct {
  host string
  port uint16
  password string
  db int
}

func main() {
  var mysqlConfig MySQLConfig
  mysqlConfig.host = "172.20.0.1"
  mysqlConfig.port = 3306
  mysqlConfig.user = "goserver"
  mysqlConfig.password = "goserver"

  var err error
  fmt.Printf(
    "%s:%s@tcp(%s:%d)/%s\n",
    mysqlConfig.user,
    mysqlConfig.password,
    mysqlConfig.host,
    mysqlConfig.port,
    "login")
  loginDB, err = sql.Open(
    "mysql",
    fmt.Sprintf(
      "%s:%s@tcp(%s:%d)/%s",
      mysqlConfig.user,
      mysqlConfig.password,
      mysqlConfig.host,
      mysqlConfig.port,
      "login"))
  if err != nil {
    panic(err.Error())
  }

  err = loginDB.Ping()
  if err != nil {
    panic(err.Error())
  }

  defer loginDB.Close()

  ctx = context.Background()

  var redisConfig RedisConfig
  redisConfig.host = "172.20.0.1"
  redisConfig.port = 6379
  redisConfig.password = "root"
  redisConfig.db = 0

  sessionDB = redis.NewClient(&redis.Options {
    Addr: fmt.Sprintf("%s:%d", redisConfig.host, redisConfig.port),
    Password: redisConfig.password,
    DB: redisConfig.db})

  _, err = sessionDB.Ping(ctx).Result()
  if err != nil {
    panic(err.Error())
  }

  http.HandleFunc("/", login)
  http.HandleFunc("/signup", signup)
  http.HandleFunc("/logout", logout)
  http.HandleFunc("/lobby", lobby)

  http.ListenAndServe(":80", nil)
}

func login(w http.ResponseWriter, req *http.Request) {
  if req.Method == http.MethodPost {
    id := req.PostFormValue("id")
    password := req.PostFormValue("password")

    user, err := GetUser(loginDB, id)
    if err != nil {
      return
    }
    if password == user.password {
      var session Session
      session.sessionId = uuid.New().String()
      session.userId = user.id
      session.created = time.Now().Format(time.StampNano)

      CreateSession(sessionDB, ctx, session)

      cookie := &http.Cookie {
        Name: "session",
        Value: session.sessionId}

      http.SetCookie(w, cookie)
      http.Redirect(w, req, "/lobby", http.StatusSeeOther)
      return
    }
  }

  http.ServeFile(w, req, "./web/template/login.html")
}


func signup(w http.ResponseWriter, req *http.Request) {
  if req.Method == http.MethodPost {
    id := req.PostFormValue("id")
    name := req.PostFormValue("name")
    password := req.PostFormValue("password")
    check := req.PostFormValue("check")

    if password == check {
      user := User{id: id, name: name, password: password}
      err := CreateUser(loginDB, user)

      if err == nil {
        http.Redirect(w, req, "/", http.StatusSeeOther)
        return
      }
    }
  }

  http.ServeFile(w, req, "./web/template/signup.html")
}

func logout(w http.ResponseWriter, req *http.Request) {
  cookie, _ := req.Cookie("session")
  RemoveSession(sessionDB, ctx, cookie.Value)

  cookie = &http.Cookie {
    Name: "session",
    Value: "",
    MaxAge: -1}
  http.SetCookie(w, cookie)
  http.Redirect(w, req, "/", http.StatusSeeOther)
}

func lobby(w http.ResponseWriter, req *http.Request) {
  http.ServeFile(w, req, "./web/template/lobby.html")
}
