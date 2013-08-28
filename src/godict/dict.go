/* Key-value data structure with non-blocking resize */
package godict

import (
	"fmt"
	mmh "murmur3"
	"sync"
)

const hash_seed uint32 = 6012
const perturb_shift = 5
const resize_ratio = 1.5

func GenHash(key string) uint32 {
	return mmh.MurMur3_32([]byte(key), hash_seed)
}

func New() *Dict {
	d := new(Dict)
	d.dict = make([]entry, 8, 8)
	d.mask = 7
	return d
}

type hashTable []entry

type Dict struct {
	sync.RWMutex
	active    uint32
	dict      hashTable
	sparedict hashTable
	mask      uint32 // mask = size - 1
	sparemask uint32
	rehashing bool
}

func (d *Dict) Active() uint32 {
	return d.active
}

//Set sets string value to key, spawn rehashing if needed
func (d *Dict) Set(key, value string) error {

	hash := GenHash(key)

	d.Lock()
	slot, err := d.lookUpEntry(key, hash)

	if err != nil {
		return err
	}

	slot.Init(key, value, hash)
	d.active++
	d.Unlock()

	d.resizeIfNeeded()

	return nil
}

//Get retrieve slot from dict, spawn error if no key in dict
func (d *Dict) Get(key string) (slot *entry, err error) {
	hash := GenHash(key)

	d.RLock()
	defer d.RUnlock()
	slot, err = d.lookUpEntry(key, hash)

	if err == nil && slot.data == nil {
		err = fmt.Errorf("Key %v missing in the dictionary", key)
	}

	return
}

//Delete mark slot as deleted and wipe it`s data, spawn error if no key in dict
func (d *Dict) Delete(key string) error {
	hash := GenHash(key)

	d.Lock()
	defer d.Unlock()
	slot, err := d.lookUpEntry(key, hash)

	if err != nil {
		return err
	}

	if slot.data == nil {
		return fmt.Errorf("Key %v missing in the dictionary", key)
	}

	slot.delete()
	d.active--

	return nil
}

// Look for entry by key and hash in hashtable, returns pointer to entry
func (ht hashTable) findSlot(key string, hash, mask uint32) *entry {

	var freeSlot *entry

	index := hash & mask
	slot := &ht[index]

	if slot.deleted {
		freeSlot = slot
	} else {
		if slot.data == nil || slot.key == key {
			return slot
		}
	}

	for perturb := hash; ; perturb >>= perturb_shift {
		index = ((index << 2) + index + perturb + 1) & mask
		slot = &ht[index]

		if slot.deleted {
			if freeSlot == nil {
				freeSlot = slot
			}
			continue
		}

		if slot.data == nil {
			if freeSlot != nil {
				slot = freeSlot
			}
			return slot
		}

		if slot.key == key {
			return slot
		}
	}
}

func (d *Dict) lookUpEntry(key string, hash uint32) (*entry, error) {
	slot := d.dict.findSlot(key, hash, d.mask)

	if slot == nil {
		return nil, fmt.Errorf("Not slot found for key %v, hash %v", key, hash)
	}

	if slot.rehashed {
		slot = d.sparedict.findSlot(key, hash, d.sparemask)
	}

	return slot, nil
}

// rehash make incremental rehashing to sparedict
func (d *Dict) rehash(newsize uint32) {

	d.sparedict = make([]entry, newsize, newsize)
	d.sparemask = newsize - 1

	for _, e := range d.dict {
		d.Lock()
		if e.data != nil {
			slot := d.sparedict.findSlot(e.key, e.hash, d.sparemask)
			slot.Init(e.key, e.value, e.hash)
		}
		e.rehashed = true
		d.Unlock()
	}
	d.Lock()
	defer d.Unlock()
	d.mask = d.sparemask
	d.dict = d.sparedict
	d.rehashing = false
	d.sparedict = nil
	d.sparemask = 0
}

// isReadyForResize atomically check if we can resize and set rehashing true
// in this case
//
// returns true if we must begin resize or false if we must not
func (d *Dict) isReadyForResize() bool {
	d.Lock()
	defer d.Unlock()
	if float64(d.mask+1)/float64(d.active) >= resize_ratio || d.rehashing {
		return false
	}
	d.rehashing = true
	return true
}

func (d *Dict) resizeIfNeeded() {

	if !d.isReadyForResize() {
		return
	}

	newsize := d.mask + 1
	active := d.active

	var mul uint32

	switch {
	case newsize < 50000:
		mul = 2
	default:
		mul = 4
	}

	for ; newsize <= mul*active; newsize <<= 1 {
	}
	d.rehash(newsize)
}
