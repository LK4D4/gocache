package main

import (
	"bufio"
	"clparse"
	"fmt"
	log "logging"
	"net"
)

func processTcpInput(input string) (string, error) {
	command, argString := clparse.SplitCommand(input)
	opts, ok := commandsMap[command]
	if !ok {
		return "", commandErr{fmt.Sprintf("Wrong command %s", command)}
	}
	args, err := clparse.ParseArgs(argString, opts.argNumber)
	if err != nil {
		return "", commandErr{err.Error()}
	}
	return opts.f(args...), nil
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	defer log.Debug("Connection closed: %v", conn.RemoteAddr())
	log.Debug("Incomming connection: %v", conn.RemoteAddr())
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		input := scanner.Text()
		log.Debug("Incomming command: %s", input)
		res, err := processTcpInput(input)
		if err != nil {
			fmt.Fprintln(conn, err)
			continue
		}
		fmt.Fprintln(conn, res)
	}
}

func runServer(host string, port int) {
	addr := fmt.Sprintf("%s:%d", host, port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Err("%v", err)
	}
	log.Info("Tcp listener running on %v", addr)
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Err("%v", err)
			continue
		}
		go handleConnection(conn)
	}
}
