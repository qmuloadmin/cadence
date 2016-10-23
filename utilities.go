package cadence

import (
	"encoding/json"
	"github.com/nu7hatch/gouuid"
	"time"
	"os"
	"fmt"
	"io"
)

func unmarshallJSONRequest(json_string []byte) (Request, error) {
	request := Request{}
	err := json.Unmarshal(json_string, &request)
	return request, err
}

func newUUID() uuid.UUID {
	id, err := uuid.NewV4()
	if err != nil {
		panic(err)
	}
	return *id
}

func isToday(self time.Time) bool {
	if self.Year() == time.Now().Year() &&
		self.Month() == time.Now().Month() &&
		self.Day() == time.Now().Day() {
		return true
	}
	return false
}

func secondsToGo(self time.Time) int64 {
	return self.Unix() - time.Now().Unix()
}

func taskListContains(self []task, str string) bool {
	for _, item := range self {
		if item.id == str {
			return true
		}
	}
	return false
}

func log(caller string, message string) {
	f, err := os.Open(localConf.log_file)
	if err != nil {
		switch err.(type) {
		case *os.PathError:
			// If a PathError is the problem, the file probably doesn't exist. Create it.
			f, err = os.Open(localConf.log_file)
		default:
			fmt.Println(err.Error())
		}
	}
	defer f.Close()
	message = "[" + time.Now().String() + "] " + caller + " | " + message
	f.WriteString(message + "\n")
	// Also print so we get 'tee' style logging
	fmt.Println(message)
}

func readToString(reader io.ReadCloser) (string, error) {
	finalBuffer := make([]byte, 100)
	for {
		buffer := make([]byte, 100)
		_, err := reader.Read(buffer)
		if err != nil {
			if err != io.EOF {
				return string(finalBuffer), err
			}
			finalBuffer = append(finalBuffer, buffer...)
			return string(finalBuffer), nil
		}
		finalBuffer = append(finalBuffer, buffer...)
	}
}