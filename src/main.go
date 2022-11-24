package main

import (
  "database/sql"
  "net/http"
  "fmt"
  "time"
  _ "github.com/go-sql-driver/mysql"
  "context"
  "github.com/go-redis/redis/v9"
)

var loginDB *sql.DB
var ctx context.Context
var redisClient *redis.Client

var hub = newHub()

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
  mysqlConfig.host = "192.168.240.1"
  mysqlConfig.port = 3306
  mysqlConfig.user = "goserver"
  mysqlConfig.password = "goserver"

  var err error
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
  redisConfig.host = "192.168.240.1"
  redisConfig.port = 6379
  redisConfig.password = "root"
  redisConfig.db = 0

  redisClient = redis.NewClient(&redis.Options {
    Addr: fmt.Sprintf("%s:%d", redisConfig.host, redisConfig.port),
    Password: redisConfig.password,
    DB: redisConfig.db})

  _, err = redisClient.Ping(ctx).Result()
  if err != nil {
    panic(err.Error())
  }

  go hub.run()

  http.HandleFunc("/", login)
  http.HandleFunc("/signup", signup)
  http.HandleFunc("/logout", logout)
  http.HandleFunc("/lobby", lobby)
  http.HandleFunc("/chat", chat)
  http.HandleFunc("/deleteAccount", deleteAccount)
  http.Handle("/web/", http.StripPrefix("/web/", http.FileServer(http.Dir("web"))))

  http.ListenAndServe(":80", nil)
}

func resetSessionCookie(w http.ResponseWriter) {
  cookie := &http.Cookie {
    Name: "session",
    Value: "",
    MaxAge: -1}
  http.SetCookie(w, cookie)
}

func login(w http.ResponseWriter, req *http.Request) {
  if req.Method == http.MethodGet {
    cookie, err := req.Cookie("session")
    if err != http.ErrNoCookie {
      if IsValidSession(redisClient, ctx, cookie.Value) {
        http.Redirect(w, req, "/lobby", http.StatusSeeOther)
        return
      }
    }
  }

  if req.Method == http.MethodPost {
    id := req.PostFormValue("id")
    password := req.PostFormValue("password")

    user, err := GetUser(loginDB, id)
    if err == nil {
      if password == user.password {
        var session Session
        session.userId = user.id
        session.created = time.Now().Format(time.StampNano)

        session, err = CreateSession(redisClient, ctx, session)
        if err == nil {
          cookie := &http.Cookie {
            Name: "session",
            Value: session.sessionId,
          }
          http.SetCookie(w, cookie)
        }

        http.Redirect(w, req, "/lobby", http.StatusSeeOther)
        return
      }
    }
  }

  http.ServeFile(w, req, "./web/login.html")
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

  http.ServeFile(w, req, "./web/signup.html")
}

func logout(w http.ResponseWriter, req *http.Request) {
  cookie, err := req.Cookie("session")
  if err != http.ErrNoCookie {
    RemoveSession(redisClient, ctx, cookie.Value)
  }

  http.Redirect(w, req, "/", http.StatusSeeOther)
}

func lobby(w http.ResponseWriter, req *http.Request) {
  cookie, err := req.Cookie("session")
  if err == http.ErrNoCookie || !IsValidSession(redisClient, ctx, cookie.Value) {
    resetSessionCookie(w)
    http.Redirect(w, req, "/", http.StatusSeeOther)
    return
  }

  http.ServeFile(w, req, "./web/lobby.html")
}

func deleteAccount(w http.ResponseWriter, req *http.Request) {
  cookie, err := req.Cookie("session")
  if err == http.ErrNoCookie || !IsValidSession(redisClient, ctx, cookie.Value) {
    resetSessionCookie(w)
    http.Redirect(w, req, "/", http.StatusSeeOther)
  }

  if req.Method == http.MethodPost {
    cookie, err := req.Cookie("session")
    if err != nil {
      resetSessionCookie(w)
      http.Redirect(w, req, "/", http.StatusSeeOther)
      return
    }

    session, err := GetSession(redisClient, ctx, cookie.Value)
    if err != nil {
      resetSessionCookie(w)
      http.Redirect(w, req, "/", http.StatusSeeOther)
      return
    }

    user, err := GetUser(loginDB, session.userId)
    if err != nil {
      resetSessionCookie(w)
      http.Redirect(w, req, "/", http.StatusSeeOther)
      return
    }

    password := req.PostFormValue("password")
    retype := req.PostFormValue("retype")

    if password == retype && user.password == password {
      RemoveUser(loginDB, session.userId)
      
      resetSessionCookie(w)
      http.Redirect(w, req, "/", http.StatusSeeOther)
      return
    }
  }

  http.ServeFile(w, req, "./web/deleteAccount.html")
}

func chat(w http.ResponseWriter, req *http.Request) {
  conn, err := upgrader.Upgrade(w, req, nil)
  if err != nil {
    return
  }

  cookie, err := req.Cookie("session")
  if err != nil {
    return
  }

  session, err := GetSession(redisClient, ctx, cookie.Value)
  if err != nil {
    return
  }

  client := &Client {
    hub: hub,
    conn: conn,
    userId: session.userId,
    send: make(chan []byte, 256),
  }
  client.hub.register <- client

  go client.writePump()
  go client.readPump()
}
