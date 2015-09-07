package main

import (
	"time"
)

type User struct {
	Email string
	StripeId string
	Notifications bool
	Active time.Time
}
