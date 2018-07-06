package vars

import (
	"bytes"
	"fmt"
	"io/ioutil"
)

var (
	bootVarPrefix = []byte("direktil.var.")
)

func BootArgs() [][]byte {
	ba, err := ioutil.ReadFile("/proc/cmdline")
	if err != nil {
		// should not happen
		panic(fmt.Errorf("failed to read /proc/cmdline: ", err))
	}

	return bytes.Split(ba, []byte{' '})
}

func BootArgValue(prefix, defaultValue string) string {
	prefixB := []byte("direktil." + prefix + "=")
	for _, ba := range BootArgs() {
		if bytes.HasPrefix(ba, prefixB) {
			return string(ba[len(prefixB):])
		}
	}

	return defaultValue
}
