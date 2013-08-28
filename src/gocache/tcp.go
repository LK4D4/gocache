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
		trimmedString := scanner.Text()
		log.Debug("Incomming command: %s", trimmedString)
		res, err := processTcpInput(trimmedString)
		if err != nil {
			fmt.Fprintf(conn, "%v\n", err)
			continue
		}
		fmt.Fprintf(conn, "%s\n", res)
	}
}

func runServer(port int) {
	addr := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Err(err.Error())
	}
	log.Info("Tcp listener running on %v", addr)
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Err(err.Error())
			continue
		}
		go handleConnection(conn)
	}
}
