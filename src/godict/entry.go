package godict

import (
	"time"
)

// Data for dictionary entry
type data struct {
	key   string
	value string
	hash  uint32
}

// entry of dictionary
type entry struct {
	*data
	time.Time
	rehashed bool // if rehashed when rehashing in progress
	deleted  bool // if was used and then deleted
	expire   time.Duration
}

// newData creates empty Data structure
func newData(key, value string, hash uint32) *data {
	return &data{key, value, hash}
}

// init rewrite entry with specified Data
func (e *entry) init(key, value string, hash uint32) {
	e.data = newData(key, value, hash)
	e.Time = time.Now()
	e.deleted = false
}

func (e *entry) access() {
	e.Time = time.Now()
}

func (e *entry) setExpire(sec uint32) {
	e.expire = time.Duration(sec) * time.Second
}

func (e *entry) Value() string {
	return e.value
}

func (e *entry) Deleted() bool {
	if e.deleted {
		return true
	}
	if e.expire != 0 && time.Since(e.Time) > e.expire {
		e.delete()
		return true
	}
	return false
}

func (e *entry) delete() {
	e.data = nil
	e.deleted = true
	e.expire = 0
}
