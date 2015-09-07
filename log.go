package main

import (
	"time"
)

type Log struct {
	Domain string
	Path string
	Size int64
	Time time.Time
}

func (l Log) Save() {

}
