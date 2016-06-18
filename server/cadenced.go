// cadenced.go
package main

import (
	"cadence"
	"fmt"
)

func main() {
	client_msgs := cadence.Start()
	fmt.Println("Now waiting for events")
	for {
		message := <-client_msgs
		switch {
		case message == cadence.DIRECTIVE_SHUTDOWN:
			fmt.Println("Shutting down now")
			break DONE
		case message == cadence.DIRECTIVE_RELOAD_CONFIG:
			fmt.Println("Reloading configuration and handling any changes...")
			cadence.Reload()
		}
	}
DONE:
}
