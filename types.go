package cadence

import (
	"time"
)

/*
tasks define a collection of commands to be run at a particular time.
task commands are always run synchronously. tasks themselves (if two are
scheduled for the same time) are run asynchronously. A task may be provided
a parent UID of another task to ensure that task has completed successfully
via the 'parent' attribute. tasks will start soon after their start time, and
continue to run until their timeout value.
*/

type task struct {
	id           string
	start        time.Time
	end          time.Time
	commands     []string
	tests        []test
	recovery     uint8
	r_value      string
	timeout      int
	dependencies []*task
	last_state   bool
	is_old       bool
	persist      bool
}

type test struct {
	method       uint8
	value        string
	fail_on_true bool
}

// type command is a target for JSON Unmarshalling for client commands that aren't tasks
type Command struct {
	Cmd   string
	Value string
}

// Request types are targets for JSON unmarshalling

type New_Task_Request struct {
	Start      string
	End        string
	Commands   []string
	Tests      []Test_Request
	Recovery   string
	R_value    string
	Timeout    string
	Dependency string
	Persist    string
}

type Test_Request struct {
	Method       string
	Value        string
	Fail_on_true string
}

type Request struct {
	Command  *Command
	New_task *New_Task_Request
}

type Zone struct {
	hosts      []string
	port       uint16
	automonous bool
}

type Conf struct {
	zones       []Zone
	self        string
	client_port uint16
	host_port   uint16
	autonomous  bool
	my_zone     *Zone
	log_file    string
	dirty       bool
	environment map[string]string
}
