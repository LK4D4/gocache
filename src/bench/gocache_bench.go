/* Benchmark for gocache */

package main

import (
	"bufio"
	"flag"
	"fmt"
	log "logging"
	"math/rand"
	"net"
	"strings"
	"sync/atomic"
	"time"
)

const chars = "abcdefghijklmnopqrstuvwxyz" +
	"ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
	"~!@#$%^&*()-_+={}[]\\|<,>.?/\"';:` "

func init() {
	rand.Seed(time.Now().UnixNano())
}

func randomString(length int) string {

	res := make([]byte, length)

	for i := 0; i < length; i++ {
		res[i] = chars[rand.Intn(len(chars))]
	}

	return string(res)
}

func genTestTable(length int) map[string]string {
	res := make(map[string]string)

	for i := 0; i < length; i++ {
		keyLen := rand.Intn(5) + 5
		dataLen := rand.Intn(50) + 50
		key := randomString(keyLen)
		data := randomString(dataLen)
		res[key] = data
	}
	return res

}

var quit = make(chan bool, 1)

var setted = make(chan string, 1000)
var deleted = make(chan string, 1000)
var saved = make(chan string, 1000)

var ops uint32 = 0

func Setter(connString string, testTable map[string]string) {
	defer close(setted)
	conn, err := net.Dial("tcp", connString)
	if err != nil {
		log.Err("Error on connection: %v", err)
	}
	defer conn.Close()

	reader := bufio.NewReader(conn)
	for key, value := range testTable {
		fmt.Fprintf(conn, "set %q %q\n", key, value)
		res, _ := reader.ReadString('\n')
		res = strings.TrimSpace(res)
		if res != "OK" {
			log.Err("Answer must be OK, receive %s", res)
			panic(res)
		}
		log.Debug("Answer: %s", res)
		atomic.AddUint32(&ops, 1)
		setted <- key
	}
	log.Info("Setter done successfully")
}

func Deleter(connString string, testTable map[string]string) {
	defer close(saved)
	defer close(deleted)
	conn, err := net.Dial("tcp", connString)
	if err != nil {
		log.Err("Error on connection: %v", err)
	}
	defer conn.Close()

	reader := bufio.NewReader(conn)
	for key := range setted {
		if rand.Float32() > 0.3 {
			saved <- key
			continue
		}
		fmt.Fprintf(conn, "delete %q\n", key)
		res, _ := reader.ReadString('\n')
		res = strings.TrimSpace(res)
		if res != "OK" {
			log.Err("Answer must be OK, receive %s", res)
			panic(res)
		}
		atomic.AddUint32(&ops, 1)
		deleted <- key
	}
	log.Info("Deleter done successfully")
}

func OkGetter(connString string, testTable map[string]string) {
	defer func() {
		quit <- true
	}()
	conn, err := net.Dial("tcp", connString)
	if err != nil {
		log.Err("Error on connection: %v", err)
	}
	defer conn.Close()

	reader := bufio.NewReader(conn)
	for key := range saved {
		fmt.Fprintf(conn, "get %q\n", key)
		answer, _ := reader.ReadString('\n')
		res := strings.SplitN(answer, " ", 2)
		retVal := strings.TrimSuffix(res[1], "\n")
		if res[0] != "OK" {
			log.Err("Error from gocache when get key %q: %v", key, retVal)
			panic(res)
		}
		value := testTable[key]
		if retVal != value {
			log.Err("Error, key %q contains %q, but must %q", key, res, value)
			panic(res)
		}
		atomic.AddUint32(&ops, 1)
		log.Debug("Answer: %s", res)
	}
	log.Info("OkGetter done successfully")
}

func ErrGetter(connString string, testTable map[string]string) {
	defer func() { quit <- true }()
	conn, err := net.Dial("tcp", connString)
	if err != nil {
		log.Err("Error on connection: %v", err)
	}
	defer conn.Close()

	reader := bufio.NewReader(conn)
	for key := range deleted {
		fmt.Fprintf(conn, "get %q\n", key)
		answer, _ := reader.ReadString('\n')
		res := strings.SplitN(answer, " ", 2)
		retVal := strings.TrimSuffix(res[1], "\n")
		if res[0] != "ERR" {
			log.Err("OK from gocache, but we delete key %q, ans: %v", key, retVal)
			panic(res)
		}
		atomic.AddUint32(&ops, 1)
	}
	log.Info("ErrGetter done successfully")
}

func main() {
	var count = flag.Int("c", 10000, "Size of testing table")
	var verbose = flag.Int("v", 4, "Logging verbosity")
	var host = flag.String("host", "", "Gocache host")
	var port = flag.Int("port", 6090, "Gocache port")
	flag.Parse()

	connString := fmt.Sprintf("%v:%v", *host, *port)
	log.SetVerbosity(*verbose)
	log.Info("Starting gocache benchmark")

	testTable := genTestTable(*count)

	var startTime time.Time
	var dur float64

	startTime = time.Now()

	go Setter(connString, testTable)
	go Deleter(connString, testTable)
	go OkGetter(connString, testTable)
	go ErrGetter(connString, testTable)

	<-quit
	<-quit

	dur = time.Since(startTime).Seconds()
	log.Info("Performed %d operations in %5fs, %.2f hits per second", ops, dur, float64(*count*2)/dur)
}
