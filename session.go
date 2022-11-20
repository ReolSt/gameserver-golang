package main

import (
  "context"
  "github.com/go-redis/redis/v9"
)

type Session struct {
  sessionId string
  userId string
  created string
}

func CreateSession(
  client *redis.Client,
  ctx context.Context,
  session Session) {

}

func RemoveSession(
  client *redis.Client,
  ctx context.Context,
  sessionId string) {

}
