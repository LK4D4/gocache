package godict

import (
	"math/rand"
	"runtime"
	"testing"
	"time"
)

const chars = "abcdefghijklmnopqrstuvwxyz" +
	"ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
	"~!@#$%^&*()-_+={}[]\\|<,>.?/\"';:` "

func randomString(length int) string {

	res := make([]byte, length)

	for i := 0; i < length; i++ {
		res[i] = chars[rand.Intn(len(chars))]
	}

	return string(res)
}

func isEmpty(e entry, t *testing.T) {
	if e.data != nil {
		t.Error("Element is not nil")
	}
}

//Test_CreateNew tests parameters of new dictionary
func Test_CreateNew(t *testing.T) {
	d := New()

	if d.active != 0 {
		t.Error("There is not 0 active elements count in new dict")
	}

	if d.mask != 7 {
		t.Error("Mask of new dict is not equal 7")
	}

	if len(d.dict) != 8 {
		t.Error("Size of new dict is not equal 8")
	}

	if d.sparedict != nil {
		t.Error("Spare dict is not nil")
	}

	if cap(d.dict) != 8 {
		t.Error("Cap of new dict is not equal 8")
	}
	for _, e := range d.dict {
		isEmpty(e, t)
	}
}

//Test_Set tests setting elements in dict
func Test_Set(t *testing.T) {
	d := New()

	d.Set("a", "1")

	if d.active != 1 {
		t.Errorf("Wrong number of active slots: %v, must be 1", d.active)
	}

	var elts []entry

	for _, e := range d.dict {
		if e.data != nil && e.key == "a" {
			elts = append(elts, e)
		}
	}

	if len(elts) == 0 {
		t.Error("Element not found in dict")
	}

	if len(elts) > 1 {
		t.Error("Element found but in multiple slots")
	}

	elt := elts[0]

	if elt.value != "1" {
		t.Error("Element found, but value is not equal 1")
	}

	hsh := GenHash("a")

	if elt.hash != hsh {
		t.Error("Element found but with wrong hash")
	}

	if elt.deleted {
		t.Error("Element found but already marked as deleted")
	}
}

func BenchmarkSet(b *testing.B) {
	testTable := make([]string, 10000)
	for i := 0; i < len(testTable); i++ {
		testTable[i] = randomString(3)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		d := New()
		for _, elt := range testTable {
			d.Set(elt, "1")
		}
	}
}

//Test_Set tests setting elements in dict and resize on 2/3 fill
func Test_SetAndResize(t *testing.T) {
	runtime.GOMAXPROCS(2)

	d := New()

	test_keys1 := map[string]string{
		"a": "1",
		"b": "2",
		"c": "3",
		"d": "4",
		"e": "5",
		"f": "6",
	}

	for key, value := range test_keys1 {
		err := d.Set(key, value)
		if err != nil {
			t.Errorf("Error %v while inserting key %v", err, key)
		}
	}

	if d.active != uint32(len(test_keys1)) {
		t.Errorf("Wrong number of active slots: %v, must be %v",
			d.active, len(test_keys1))
	}

	for key, value := range test_keys1 {
		res, err := d.Get(key)
		if err != nil {
			t.Errorf("Error %v on retrieving key %v", err, key)
		}
		if res.value != value {
			t.Errorf("Wrong value %v for key %v", res, key)
		}
	}

	// waiting for rehash, cause of deadlock
	time.Sleep(time.Second / 10)

	if d.mask != 15 {
		t.Errorf("Wrong mask %v after resize, must be 15", d.mask)
	}

	if len(d.dict) != 16 {
		t.Errorf("Wrong len of resized dict: %v, must be 16", len(d.dict))
	}

	if cap(d.dict) != 16 {
		t.Errorf("Wrong cap of resized dict: %v, must be 16", cap(d.dict))
	}

	test_keys2 := map[string]string{
		"g": "7",
		"h": "8",
		"i": "9",
		"j": "10",
		"k": "11",
	}

	for key, value := range test_keys2 {
		err := d.Set(key, value)
		if err != nil {
			t.Errorf("Error %v while inserting key %v", err, key)
		}
	}

	if d.active != uint32(len(test_keys1)+len(test_keys2)) {
		t.Errorf("Wrong number of active slots: %v, must be %v",
			d.active, len(test_keys1)+len(test_keys2))
	}

	for key, value := range test_keys1 {
		res, err := d.Get(key)
		if err != nil {
			t.Errorf("Error %v on retrieving key %v", err, key)
		}
		if res.value != value {
			t.Errorf("Wrong value %v for key %v", res, key)
		}
	}

	for key, value := range test_keys2 {
		res, err := d.Get(key)
		if err != nil {
			t.Errorf("Error %v on retrieving key %v", err, key)
		}
		if res.value != value {
			t.Errorf("Wrong value %v for key %v", res, key)
		}
	}

	time.Sleep(time.Second / 10)

	if d.mask != 31 {
		t.Errorf("Wrong mask %v after resize, must be 31", d.mask)
	}

	if len(d.dict) != 32 {
		t.Errorf("Wrong len of resized dict: %v, must be 32", len(d.dict))
	}

	if cap(d.dict) != 32 {
		t.Errorf("Wrong cap of resized dict: %v, must be 32", cap(d.dict))
	}
}

//Test_Get tests getting element from dict
func Test_Get(t *testing.T) {
	d := New()

	d.Set("a", "1")

	res, err := d.Get("a")

	if err != nil {
		t.Errorf("Get failed with error %v", err)
	}

	if res.value != "1" {
		t.Errorf("Get wrong value %v", res)
	}
}

//Test_GetUnexisting tests getting element from dict without setting it first
func Test_GetUnexisting(t *testing.T) {
	d := New()

	slot, err := d.Get("a")

	if err == nil {
		t.Logf("Get did not fail. slot = %v. What value he returned?", slot)
		v := slot.Value()
		t.Errorf("Get did not fail and returned %v", v)
	}

}

//Test_Delete tests deleting element from dict
func Test_Delete(t *testing.T) {
	d := New()

	d.Set("a", "1")

	err := d.Delete("a")

	if err != nil {
		t.Errorf("Delete failed with error %v", err)
	}

	slot, err := d.Get("a")

	if err == nil {
		t.Logf("Get did not fail after delete. slot = %v. What value he returned?", slot)
		v := slot.Value()
		t.Errorf("Get did not fail and returned %v", v)
	}

	d.Set("a", "1")

	slot, err = d.Get("a")

	if err != nil {
		t.Errorf("Get failed with error %v", err)
	}

	if slot.value != "1" {
		t.Errorf("Get wrong value %v", slot.value)
	}

	if slot.deleted {
		t.Errorf("Key was set again, but still deleted, slot %v", slot)
	}
}

//Test_DeleteUnexisting tests delete element from dict without setting it first
func Test_DeleteUnexisting(t *testing.T) {
	d := New()

	err := d.Delete("a")

	if err == nil {
		t.Errorf("Delete did not fail. dict: %v", d)
	}

}
