package cadence

// Define constants for determining actions to take on task failure

const ACTION_RERUN_TASK uint8 = 0
const ACTION_FAIL uint8 = 1
const ACTION_BATCH uint8 = 2
const ACTION_RERUN_TASK_DELAY uint8 = 3
const ACTION_RERUN_TASK_DIFFERENT_HOST uint8 = 4
const ACTION_RERUN_FAILED uint8 = 5

// Define constants for determining methodologies of validating tasks

const METHOD_EXIT_CODE uint8 = 0
const METHOD_STDOUT_MATCH uint8 = 1
const METHOD_STDOUT_EMPTY = 2
const METHOD_STDERR_MATCH uint8 = 3
const METHOD_STDERR_EMPTY uint8 = 4
const METHOD_TIMEOUT uint8 = 5

// Define client directives that require logic in main loop

const DIRECTIVE_RELOAD_CONFIG uint8 = 0
const DIRECTIVE_SHUTDOWN uint8 = 1
const DIRECTIVE_UNKNOWN uint8 = 255

const CONF_FILE = "/etc/cadence.conf"
