package cadence

// Define constants for determining actions to take on task failure

const ACTION_RERUN_TASK uint8 = 0                // rerun the entire task (all commands) immediately
const ACTION_BATCH uint8 = 2                     // rerun the entire task as batch job
const ACTION_RERUN_TASK_DELAY uint8 = 3          // rerun the entire task after a delay
const ACTION_RERUN_TASK_DIFFERENT_HOST uint8 = 4 // rerun the entire task on a different host
const ACTION_RERUN_FAILED uint8 = 5              // rerun only failed commands immediately
const ACTION_NONE uint8 = 6                      // take no action, task fails immediately

// Define constants for determining methodologies of validating tasks

const METHOD_EXIT_CODE uint8 = 0
const METHOD_STDOUT_MATCH uint8 = 1
const METHOD_STDOUT_EMPTY = 2
const METHOD_STDERR_MATCH uint8 = 3
const METHOD_STDERR_EMPTY uint8 = 4
const METHOD_NONE uint8 = 6

// Define states for a given task

const STATE_PENDING_START uint8 = 0
const STATE_PENDING_DEPS uint8 = 1
const STATE_CANCELLED uint8 = 2
const STATE_RUNNING uint8 = 3
const STATE_FAILED_TEST uint8 = 4
const STATE_FAILED_TIMEOUT uint8 = 5
const STATE_SUCCESSFUL uint8 = 6

// Define client directives that require logic in main loop

const DIRECTIVE_RELOAD_CONFIG uint8 = 0
const DIRECTIVE_SHUTDOWN uint8 = 1
const DIRECTIVE_UNKNOWN uint8 = 255

const CONF_FILE string = "/etc/cadence.conf"
