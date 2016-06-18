/*
cadence library contains functions for managing  the event handling of tasks,
the socket listener, and distributed transactions
*/

package cadence

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"strconv"

	"github.com/nu7hatch/gouuid"
)

var local_config Conf

// Restart the server, performing all necessary tasks and communications, then
// reload the configuration.
func Reload() bool {
	local_config = load()
	return true
}

// Start the two listeners, host and client, return channels to caller
func Start() chan uint8 {
	fmt.Print("Starting new client listener ")
	if local_config.isZero() {
		local_config = load()
	}
	fmt.Println("on port ", local_config.client_port)
	directive := make(chan uint8)
	go listen_to_client(directive)
	fmt.Println("Starting new host listener on port", local_config.host_port)
	return directive
}

// list_to_client waits on connections to the client_port, for configuration instructions
// and shutdown/restart commands, among other things.
func listen_to_client(directive chan<- uint8) {
	listener, err := net.Listen("tcp", ":"+strconv.Itoa(int(local_config.client_port)))
	if err != nil {
		panic(err)
	}
	defer listener.Close()
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println(err.Error())
		}
		reader := bufio.NewReader(conn)
		message, err := reader.ReadBytes('\n')
		fmt.Println("Message Received:", string(message))
		// Convert message from JSON into Request object and see if a command or task was provided
		request, err := unmarshall_json_request(message)
		if err != nil {
			conn.Write([]byte("Error: " + err.Error() + "\n"))
		} else {
			// Determine if a command was sent or a new task request
			fmt.Println(request)
			if request.New_task == nil {
				command := process_command(request.Command.Cmd, request.Command.Value)
			}
		}
		id := new_uuid()
		conn.Write([]byte(id.String() + "\n"))
	}
}

func process_command(cmd string, value string) uint8 {
	switch {
	case cmd == "SHUTDOWN":
		return DIRECTIVE_SHUTDOWN
	case cmd == "RELOAD":
		return DIRECTIVE_RELOAD_CONFIG
	default:
		return DIRECTIVE_UNKNOWN
	}
}

func unmarshall_json_request(json_string []byte) (Request, error) {
	request := Request{}
	err := json.Unmarshal(json_string, &request)
	return request, err
}

func new_uuid() uuid.UUID {
	id, err := uuid.NewV4()
	if err != nil {
		panic(err)
	}
	return *id
}
