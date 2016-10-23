/*
cadence library contains functions for managing  the event handling of tasks,
the socket listener, and distributed transactions
*/

package cadence

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
)

var localConf Conf
var futureTasks tasks
var todayTasks tasks

// Restart the server, performing all necessary tasks and communications, then
// reload the configuration.
func Reload() bool {
	localConf = load()
	return true
}

// Start the two listeners, host and client, return client channel to caller
func Start() chan uint8 {
	name := "Start"
	if localConf.is_zero() {
		localConf = load()
	}
	log(name, "Cadence is starting")
	log(name, "Starting new client listener on port " + strconv.Itoa(int(localConf.clientPort)))
	client := make(chan uint8)
	dispatcher := make(chan bool)
	go listenToClient(client, dispatcher)
	go exclusiveDispatch(dispatcher)
	return client
}

// listenToClient waits on connections to the client_port, for configuration instructions
// and shutdown/restart commands, among other things.
func listenToClient(directive chan<- uint8, dispatcher chan<- bool) {
	name := "ClientListener"
	listener, err := net.Listen("tcp", ":"+strconv.Itoa(int(localConf.clientPort)))
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
		log(name, "Message Received:" + string(message))
		// Convert message from JSON into Request object and see if a command or task was provided
		request, err := unmarshallJSONRequest(message)
		if err != nil {
			conn.Write([]byte("Error: " + err.Error() + "\n"))
		} else {
			// Determine if a command was sent or a new task request
			if request.New_task == nil {
				processCommand(request.Command.Cmd, request.Command.Value, directive)
			} else {
				// if New_task is not nil, then we need to handle the new task
				newID, err := processNewTaskRequest(request.New_task)
				if err != nil {
					conn.Write([]byte(err.Error() + "\n"))
					log(name, "Message Sent: " + err.Error())
				} else {
					// All went well, let the user know of the new UUID
					conn.Write([]byte("New Task ID: " + newID + "\n"))
					log(name, "New task " + newID + " created")
					dispatcher <- true
				}
			}
		}
	}
}

func processNewTaskRequest(newTaskRequest *NewTaskRequest) (string, error) {
	// function converts request to Task type, handles default values and ensures all required information is
	// provided, and if not, return the error message with details back to the client
	newTask, err := newTaskRequest.toTask()
	if err == nil {
		futureTasks.mutex.Lock()
		futureTasks.items = append(futureTasks.items, newTask)
		futureTasks.mutex.Unlock()
		// if task is today, re-update the list of tasks for today
		if isToday(newTask.start) {
			todayTasks.mutex.Lock()
			todayTasks.items = append(todayTasks.items, newTask)
			todayTasks.mutex.Unlock()
		}
		return newTask.id, nil
	} else {
		return "", err
	}
}

func exclusiveDispatch(taskListUpdate <-chan bool) {
	// Exclusive Dispatch listens for new tasks being added to today,
	// for the day to change (midnight), to individual tasks for their wake times,
	// and to the network for remote task execution mutex negotiation. It then permits
	// or denies individual tasks from executing on the local system after negotiating
	// mutex state.
	waiting := todayTasks.items
	timer := make(chan string)
	idToIndex := make(map[string]int)
	name := "Dispatcher"
	for {
		select {
		case <-taskListUpdate:
			for _, elem := range todayTasks.items {
				if !taskListContains(waiting, elem.id) {
					waiting = append(waiting, elem)
					idToIndex[elem.id] = len(waiting) - 1
					log(name, "Added " + elem.id + " to waiting list")
					go elem.sleep(timer)
				}
			}
		case id := <-timer:
			log(name, "Task " + id + " woke up")
			// since we're not yet running on distributed mutex, just fire the task
			// and it'll take care of itself.
			if taskListContains(waiting, id) {
				// We do this to make sure the task wasn't cancelled since it was
				// put to sleep. If it was, then we're just going to ignore it
				go waiting[idToIndex[id]].execute()
			}
		}
	}
}

func processCommand(cmd string, value string, client chan<- uint8) {
	switch {
	case cmd == "SHUTDOWN":
		client <- DIRECTIVE_SHUTDOWN
	case cmd == "RELOAD":
		client <- DIRECTIVE_RELOAD_CONFIG
	default:
		client <- DIRECTIVE_UNKNOWN
	}
}
