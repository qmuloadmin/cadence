package cadence

import (
	"sync"
	"time"
	"os/exec"
	"strings"
)

/*
tasks define a collection of commands to be run at a particular time.
task commands are always run synchronously. tasks themselves (if two are
scheduled for the same time) are run asynchronously. A task may be provided
a parent UID of another task to ensure that task has completed successfully
via the 'parent' attribute. tasks will start soon after their start time, and
continue to run until their timeout value.
*/

type InvalidTaskRequest struct {
	reason string
	field  string
}

func (err *InvalidTaskRequest) Error() string {
	return "ERROR: Task field '" + err.field + "' " + err.reason
}

func (err InvalidTaskRequest) isZero() bool {
	if err.reason == "" && err.field == "" {
		return true
	}
	return false
}

type InvalidTestRequest struct {
	message string
}

func (err *InvalidTestRequest) Error() string {
	return "ERROR: " + err.message
}

type tasks struct {
	mutex sync.Mutex
	items []task
}

type task struct {
	id         string    // uuid of the task
	start      time.Time // the time at which execution of the task should start
	end        time.Time // the end time, when if for some reason the job does not successfully execute before this, it should not try anymore
	commands   []string  // commands to be executed, verbatim
	tests      []test    // a list of tests to perform to determine if task was successful, from METHOD constants
	recovery   uint8     // the recovery action to take, from ACTION constants
	rValue     string    // the value associated with recovery action, if applicable
	maxRetries uint8     // the Number of times the task will retry its recovery action
	timeout    uint64    // given a single execution of task, if this time is exceeded, consider failed, attempt to recover
	dependency *task     // a list of dependent tasks that must run first, and successfully
	state      uint8     // set to values from STATE constants, indicates the state of the task
	retries    uint8     // the count of times this task has been retried
	isOld      bool      // indicates whether the job has already finished its execution cycle (regardless of state)
	persist    bool      // set to indicate that after execution life cycle, persist a record of the task
	exec_host  string    // the host the task is currently executing on, or was executed on
}

type test struct {
	method     uint8
	value      string
	failOnTrue bool
}

func (self *task) sleep(wake chan<- string) {
	duration := time.Duration(secondsToGo(self.start)) * time.Second
	time.Sleep(duration)
	wake <- self.id
}

func (self *task) execute() {
	name := "Execute ("+self.id+")"
	commandArgs := strings.Split(self.commands[0], " ")
	log(name, "Executing command " + self.commands[0])
	command := exec.Command(commandArgs[0], commandArgs[1:]...)
	stdout, err := command.StdoutPipe()
	if err != nil {
		// TODO Handle failure based on recovery property
	}
	stderr, err := command.StderrPipe()
	if err != nil {
		// TODO handle error/failure
	}
	command.Start()
	log(name, "Command executed, waiting for completion")
	// TODO find the best way to detect and kill a timed-out command
	stdoutBuffer, _ := readToString(stdout)
	stderrBuffer, _ := readToString(stderr)
	command.Wait()

	log(name, "Stdout results from command '" + stdoutBuffer + "'")
	log(name, "Stderr results from command '" + stderrBuffer + "'")
}

// type command is a target for JSON Unmarshalling for client commands that aren't tasks
type Command struct {
	Cmd   string
	Value string
}

// Request types are targets for JSON unmarshalling

type NewTaskRequest struct {
	Start      int           `required; seconds since epoch`
	End        int           `optional; if not provided, will never be too late`
	Commands   []string      `required; must include at least one command`
	Tests      []TestRequest `optional;`
	Recovery   string        `optional; defaults to ACTION_NONE`
	RValue     string        `optional;`
	Timeout    uint          `required; seconds`
	Retries    int           `optional; defaults to 1`
	Dependency string        `optional;`
	Persist    bool          `optional; defaults to false`
}

// Converts a NewTaskRequest object into a task object

func (taskRequest NewTaskRequest) toTask() (task, error) {

	newTask := task{}
	err := InvalidTaskRequest{}
	newTask.start = time.Unix(int64(taskRequest.Start), 0)
	now := time.Now()
	// Validate values are valid as provided
	if taskRequest.Start == 0 {
		err.field = "start"
		err.reason = "must be provided"
	} else if newTask.start.Before(now) {
		err.field = "start"
		err.reason = "must be in the future"
	} else if len(taskRequest.Commands) == 0 {
		err.field = "commands"
		err.reason = "must not be empty"
	} else if taskRequest.Timeout == 0 {
		err.field = "timeout"
		err.reason = "must be provided and non-zero"
	}

	if taskRequest.End != 0 {
		newTask.end = time.Unix(int64(taskRequest.End), 0)
		if newTask.end.Before(newTask.start) {
			err.field = "end"
			err.reason = "must be after start time"
		}
	}

	newTask.commands = taskRequest.Commands
	for _, req := range taskRequest.Tests {
		newTest, err := req.toTest()
		if err != nil {
			newTask.tests = append(newTask.tests, newTest)
		} else {
			return newTask, err
		}
	}
	switch taskRequest.Recovery {
	case "RERUN_TASK":
		newTask.recovery = ACTION_RERUN_TASK
	case "NONE":
		newTask.recovery = ACTION_NONE
	case "BATCH":
		newTask.recovery = ACTION_BATCH
	case "RERUN_TASK_DELAY":
		newTask.recovery = ACTION_RERUN_TASK_DELAY
	case "RERUN_TASK_DIFFERENT_HOST":
		newTask.recovery = ACTION_RERUN_TASK_DIFFERENT_HOST
	case "RERUN_FAILED":
		newTask.recovery = ACTION_RERUN_FAILED
	case "":
		newTask.recovery = ACTION_NONE
	default:
		err.field = "recovery"
		err.reason = "invalid action specified: " + taskRequest.Recovery
	}
	newTask.rValue = taskRequest.RValue
	newTask.maxRetries = uint8(taskRequest.Retries)
	newTask.timeout = uint64(taskRequest.Timeout)
	newTask.persist = taskRequest.Persist
	newTask.state = STATE_PENDING_START
	if err.isZero() {
		newID := newUUID()
		newTask.id = newID.String()
		return newTask, nil
	} else {
		return newTask, &err
	}
}

// Minimum request to generate valid task
// {"new_task":{"start":1475437451, "commands": ["foo"], "timeout":75}}

func (testRequest TestRequest) toTest() (test, error) {
	newTest := test{}
	switch testRequest.Method {
	case "EXIT_CODE":
		newTest.method = METHOD_EXIT_CODE
	case "STDOUT_MATCH":
		newTest.method = METHOD_STDOUT_MATCH
	case "STDOUT_EMPTY":
		newTest.method = METHOD_STDOUT_EMPTY
	case "STDERR_MATCH":
		newTest.method = METHOD_STDERR_MATCH
	case "STDERR_EMPTY":
		newTest.method = METHOD_STDERR_EMPTY
	case "NONE":
		newTest.method = METHOD_NONE
	default:
		err := InvalidTestRequest{}
		err.message = "Invalid test method specified: " + testRequest.Method
		return newTest, &err
	}
	newTest.value = testRequest.Value
	newTest.failOnTrue = testRequest.FailOnTrue
	return newTest, nil
}

type TestRequest struct {
	Method     string
	Value      string
	FailOnTrue bool
}

type Request struct {
	Command  *Command
	New_task *NewTaskRequest
}

type Zone struct {
	hosts      []string
	port       uint16
	autonomous bool
}

type Conf struct {
	zones       []Zone
	self        string
	clientPort  uint16
	hostPort    uint16
	autonomous  bool
	my_zone     *Zone
	log_file    string
	dirty       bool
	environment map[string]string
}
