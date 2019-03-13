package initboot

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
)

var stateFile = "/run/dkl-boot.state"

func readState() (state map[string]bool) {
	state = map[string]bool{}

	ba, err := ioutil.ReadFile(stateFile)
	if err != nil {
		if os.IsNotExist(err) {
			return
		}
		log.Fatal("failed to read state: ", err)
	}

	err = json.Unmarshal(ba, &state)
	if err != nil {
		log.Fatal("failed to parse state: ", err)
	}

	return
}

func writeState(state map[string]bool) {
	ba, err := json.Marshal(state)
	if err != nil {
		log.Fatal("failed to serialize state: ", err)
	}

	ioutil.WriteFile(stateFile, ba, 0600)
}

func step(step string, operation func()) {
	state := readState()
	if !state[step] {
		operation()

		state[step] = true
		writeState(state)
	}
}
