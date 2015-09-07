package main

type Subscription struct {
	Domain string
	StripeId string
}

func (s Subscription) Active() (active bool) {
	return
}
