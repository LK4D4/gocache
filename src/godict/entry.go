package godict

// Data for dictionary entry
type data struct {
	key   string
	value string
	hash  uint32
}

// entry of dictionary
type entry struct {
	*data
	rehashed bool // if rehashed when rehashing in progress
	deleted  bool // if was used and then deleted
}

// NewData creates empty Data structure
func newData(key, value string, hash uint32) *data {
	return &data{key, value, hash}
}

// Init rewrite entry with specified Data
func (e *entry) Init(key, value string, hash uint32) {
	e.data = newData(key, value, hash)
	e.deleted = false
}

func (e *entry) Value() string {
	return e.value
}

func (e *entry) delete() {
	e.data = nil
	e.deleted = true
}
