package main

import (
  "context"
  "fmt"
  "time"
  "errors"
  "github.com/google/uuid"
  "github.com/go-redis/redis/v9"
)

type Session struct {
  sessionId string
  userId string
  created string
  ip string
  port string
}

func FindDupliatedSession(
  client *redis.Client,
  ctx context.Context,
  userId string) (Session, error) {
  iter := client.Scan(ctx, 0, "session:*", 0).Iterator()

  var duplicatedSession Session
  for iter.Next(ctx) {
    sessionKey := iter.Val()

    values, err := client.HGetAll(ctx, sessionKey).Result()
    if err == nil {
      if values["userId"] == userId {
        duplicatedSession.sessionId = values["sessionId"]
        duplicatedSession.userId = values["userId"]
        duplicatedSession.created = values["created"]

        return duplicatedSession, err
      }
    }
  }

  return duplicatedSession, errors.New("Not Found")
}

func CreateSession(
  client *redis.Client,
  ctx context.Context,
  session Session) (Session, error) {
  duplicatedSession, err := FindDupliatedSession(client, ctx, session.userId)
  if err == nil {
    key := fmt.Sprintf("session:%s", duplicatedSession.sessionId)
    _, err = client.Expire(ctx, key, 15 * time.Minute).Result()
    RemoveSession(client, ctx, duplicatedSession.sessionId)
  }

  session.sessionId = uuid.New().String()
  key := fmt.Sprintf("session:%s", session.sessionId)

  _, err = client.HMSet(ctx, key,
    map[string]string {
      "sessionId": session.sessionId,
      "userId": session.userId,
      "created": session.created,
      "ip": session.ip,
      "port": session.port }).Result()

    if err != nil {
      return session, err
    }

    _, err = client.Expire(ctx, key, 15 * time.Minute).Result()
    return session, err
}

func GetSession(
  client *redis.Client,
  ctx context.Context,
  sessionId string) (Session, error) {
  key := fmt.Sprintf("session:%s", sessionId)
  values, err := client.HGetAll(ctx, key).Result()

  var session Session
  if err != nil {
    return session, err
  }

  session.sessionId = values["sessionId"]
  session.userId = values["userId"]
  session.created = values["created"]

  if err != nil {
    return session, err
  }

  _, err = client.Expire(ctx, key, 15 * time.Minute).Result()
  return session, err
}

func RemoveSession(
  client *redis.Client,
  ctx context.Context,
  sessionId string) error {
  key := fmt.Sprintf("session:%s", sessionId)
  _, err := client.Del(ctx, key).Result()

  return err
}

func IsValidSession(
  client *redis.Client,
  ctx context.Context,
  sessionId string) bool {
  session, err := GetSession(client, ctx, sessionId)
  if err != nil {
    return false
  }

  return session.sessionId == sessionId
}
