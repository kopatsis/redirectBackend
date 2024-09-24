package redisfn

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/stripe/stripe-go/v72/sub"
)

type UserPayment struct {
	CustomerID     string    `json:"c"`
	SubscriptionID string    `json:"s"`
	LastDate       time.Time `json:"d"`
	Active         bool      `json:"a"`
}

func GetUserPayment(rdb *redis.Client, uid string) (*UserPayment, error) {
	key := ":p:" + uid
	data, err := rdb.Get(context.Background(), key).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var p UserPayment
	err = json.Unmarshal([]byte(data), &p)
	if err != nil {
		return nil, err
	}

	return &p, nil
}

func setUserPayment(rdb *redis.Client, uid string, p *UserPayment) error {
	key := ":p:" + uid
	data, err := json.Marshal(p)
	if err != nil {
		return err
	}

	err = rdb.Set(context.Background(), key, data, 0).Err()
	if err != nil {
		return err
	}

	return nil
}

func setPaymentForUser(rdb *redis.Client, uid, subID string) error {
	return rdb.Set(context.Background(), ":b:"+subID, uid, 0).Err()
}

func GetUserBySubID(rdb *redis.Client, subID string) (string, error) {
	return rdb.Get(context.Background(), ":b:"+subID).Result()
}

func CheckUserPaying(rdb *redis.Client, uid string) (bool, error) {
	p, err := GetUserPayment(rdb, uid)
	if err != nil {
		return false, err
	}

	if p == nil {
		return false, nil
	}

	if !p.Active || p.CustomerID == "" || p.SubscriptionID == "" {
		return false, nil
	}

	if p.LastDate.After(time.Now()) {
		return true, nil
	}

	s, err := sub.Get(p.SubscriptionID, nil)
	if err != nil {
		return false, err
	}

	if s == nil {
		return false, errors.New("no actual active subscription")
	}

	periodEnd := time.Unix(s.CurrentPeriodEnd, 0)
	if s.Status == "active" && periodEnd.After(time.Now()) {
		p.LastDate = periodEnd
		if err := setUserPayment(rdb, uid, p); err != nil {
			return true, err
		}
		return true, nil
	}

	p.Active = false
	if err := setUserPayment(rdb, uid, p); err != nil {
		return false, err
	}

	return false, nil
}

func CreateBlankUserPayment(rdb *redis.Client, uid, custID string) error {
	p := UserPayment{CustomerID: custID}
	return setUserPayment(rdb, uid, &p)
}

func CreateSetUserPayment(rdb *redis.Client, uid, custID, subID string) error {
	p := UserPayment{CustomerID: custID, SubscriptionID: subID}
	if err := setUserPayment(rdb, uid, &p); err != nil {
		return err
	}
	return setPaymentForUser(rdb, uid, subID)
}

func SetSubOnUserPayment(rdb *redis.Client, uid, subID string) error {

	p, err := GetUserPayment(rdb, uid)
	if err != nil {
		return err
	}

	if p == nil {
		return errors.New("no user payment for this user")
	}

	p.SubscriptionID = subID
	p.Active = false

	if err := setUserPayment(rdb, uid, p); err != nil {
		return err
	}
	return setPaymentForUser(rdb, uid, subID)
}

func SetUserPaymentActive(rdb *redis.Client, uid, subID string, periodEnd time.Time) error {

	if periodEnd.Before(time.Now()) {
		return errors.New("period end date is before right now")
	}

	p, err := GetUserPayment(rdb, uid)
	if err != nil {
		return err
	}

	if p == nil {
		return errors.New("no user payment for this user")
	}

	if p.SubscriptionID != subID {
		return errors.New("incorrect subscription id for user")
	}

	p.Active = true
	p.LastDate = periodEnd

	return setUserPayment(rdb, uid, p)
}

func SetUserPaymentInactive(rdb *redis.Client, uid, subID string) error {

	p, err := GetUserPayment(rdb, uid)
	if err != nil {
		return err
	}

	if p == nil {
		return errors.New("no user payment for this user")
	}

	if p.SubscriptionID != subID {
		return errors.New("incorrect subscription id for user")
	}

	p.Active = false

	return setUserPayment(rdb, uid, p)
}
