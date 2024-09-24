package redisfn

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/go-redis/redis/v8"
)

type CookieLimit struct {
	Success   bool      `json:"s"`
	Banned    bool      `json:"b"`
	ResetDate time.Time `json:"r"`
}

func getCookieLimit(rdb *redis.Client, uid string) (CookieLimit, error) {
	key := ":c:" + uid
	data, err := rdb.Get(context.Background(), key).Result()
	if err == redis.Nil {
		return CookieLimit{}, nil
	}
	if err != nil {
		return CookieLimit{}, err
	}

	var limit CookieLimit
	err = json.Unmarshal([]byte(data), &limit)
	if err != nil {
		return CookieLimit{}, err
	}

	if !limit.Success {
		return CookieLimit{}, errors.New("not unmarshalled correclty")
	}

	return limit, nil
}

func addCookieLimit(rdb *redis.Client, uid string, limit CookieLimit) error {
	key := ":c:" + uid
	data, err := json.Marshal(limit)
	if err != nil {
		return err
	}

	err = rdb.Set(context.Background(), key, data, 0).Err()
	if err != nil {
		return err
	}

	return nil
}

func AddResetDate(rdb *redis.Client, uid string) error {
	cl, err := getCookieLimit(rdb, uid)
	if err != nil {
		return err
	}

	cl.Success = true
	cl.ResetDate = time.Now()

	return addCookieLimit(rdb, uid, cl)
}

func AddBanned(rdb *redis.Client, uid string) error {
	cl, err := getCookieLimit(rdb, uid)
	if err != nil {
		return err
	}

	cl.Success = true
	cl.Banned = true

	return addCookieLimit(rdb, uid, cl)
}

func CheckCookeLimit(rdb *redis.Client, uid string, added time.Time) (banned bool, reset bool, reterr error) {

	cl, err := getCookieLimit(rdb, uid)
	if err != nil {
		return false, false, err
	}

	if !cl.Success {
		return false, false, nil
	}

	if cl.Banned {
		return true, false, nil
	}

	if added.Before(cl.ResetDate) {
		return false, true, nil
	}

	return false, false, nil
}
