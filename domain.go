package main

import (
	"errors"
	"fmt"
	"net/http"
)

type Domain struct {
	Domain string
	Upstream string
	User string
	Subscribed bool
}

// @todo Write this.
func getDomain(domain string) (err error, d Domain) {

	d = Domain{
		Domain: "localhost",
		Upstream: "www.google.com",
		User: "",
		Subscribed: true,
	}

	return

}

func getDomainFromReq(r *http.Request) (err error, d Domain) {

	err, d = getDomain(r.URL.Host)
  return

}

func (d Domain) Save() {

}

func (d Domain) Subscribe() {

	// @todo Charge the user's account.
	d.Subscribed = true
	d.Save()

	return

}

func (d Domain) Register(u User) {

	d.User = u.Email
	d.Save()

	return

}

func (d Domain) Check() (err error) {

	// If subscribed, have unlimited transfer right now.
	if d.Subscribed {
		return
	}

	// Check logs to see if over quota.
	return errors.New(fmt.Sprintf("Domain %s is not registered", d.Domain))

}
