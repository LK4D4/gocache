/* Basic TCP server */
package tcp

import (
	"bufio"
	"commands"
	"fmt"
	log "logging"
	"net"
)

func handleConnection(conn net.Conn) {
	defer conn.Close()
	defer log.Debug("Connection closed: %v", conn.RemoteAddr())
	log.Debug("Incomming connection: %v", conn.RemoteAddr())
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		trimmedString := scanner.Text()
		log.Debug("Incomming command: %s", trimmedString)
		res, err := commands.ProcessTcpInput(trimmedString)
		if err != nil {
			fmt.Fprintf(conn, "%v\n", err)
			continue
		}
		fmt.Fprintf(conn, "%s\n", res)
	}
}

func RunServer(port int) {
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
