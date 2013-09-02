/* Key-value data structure with non-blocking resize */
package godict

import (
	"fmt"
	log "logging"
	mmh "murmur3"
	"runtime"
	"sync"
)

const (
	// constants for ratio computing
	activeMul uint32 = 3
	sizeMul   uint32 = 2

	hashSeed     uint32 = 6012
	perturbShift uint32 = 5
	rehashChunk  uint32 = 1000
)

func GenHash(key string) uint32 {
	return mmh.MurMur3_32([]byte(key), hashSeed)
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

	log.Debug("Rehashing status %v", d.rehashing)

	d.Lock()
	slot, err := d.lookUpEntry(key, hash)

	if err != nil {
		return err
	}

	slot.Init(key, value, hash)
	d.active++
	d.Unlock()

	go d.resizeIfNeeded()

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

	for perturb := hash; ; perturb >>= perturbShift {
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
		log.Debug("Slot for %q found, but it`s rehashed", key)
		slot = d.sparedict.findSlot(key, hash, d.sparemask)
	}

	return slot, nil
}

func (d *Dict) rehashChunk(l, r uint32) {
	d.Lock()
	defer d.Unlock()
	tmp := d.dict[l:r]
	for i := range tmp {
		e := &tmp[i]
		if e.data != nil {
			log.Debug("Rehashing key %q", e.key)
			slot := d.sparedict.findSlot(e.key, e.hash, d.sparemask)
			slot.data = e.data
		}
		e.rehashed = true
	}
}

// rehash make incremental rehashing to sparedict
func (d *Dict) rehash(newsize uint32) {
	log.Debug("Rehashing started")
	defer log.Debug("Rehashing finished")

	d.sparedict = make([]entry, newsize, newsize)
	d.sparemask = newsize - 1

	var lb, rb uint32 = 0, rehashChunk

	dlen := d.mask + 1

	for ; rb <= dlen; lb, rb = rb, rb+rehashChunk {
		d.rehashChunk(lb, rb)
		runtime.Gosched()
	}
	if lb != dlen {
		d.rehashChunk(lb, dlen)
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
	if ((d.mask+1)*sizeMul >= d.active*activeMul) || d.rehashing {
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
