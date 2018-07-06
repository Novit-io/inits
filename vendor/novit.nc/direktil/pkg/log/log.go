package log

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"novit.nc/direktil/pkg/color"
)

const (
	// AppendNL indicates that a forced '\n' is added.
	AppendNL byte = 1
)

var (
	logs  = map[string]*Log{}
	mutex = sync.Mutex{}

	logOutputEnabled = false
)

// Log is a log target
type Log struct {
	name string

	l           sync.Mutex
	writeToFile bool

	console io.Writer
	pending []Entry
	out     *os.File
	outTS   string
}

func Get(name string) *Log {
	mutex.Lock()
	defer mutex.Unlock()

	if log, ok := logs[name]; ok {
		return log
	}

	log := &Log{
		name:    name,
		pending: make([]Entry, 0),
	}

	if logOutputEnabled {
		log.enableFileOutput()
	}

	logs[name] = log

	return log
}

// EnableFiles flushes current logs to files, and enables output to files.
func EnableFiles() {
	mutex.Lock()
	defer mutex.Unlock()

	if logOutputEnabled {
		return
	}

	for _, log := range logs {
		// we'll let the kernel optimize, just do it all parallel
		go log.enableFileOutput()
	}

	logOutputEnabled = true
}

// DisableFiles flushes and closes current logs files, and disables output to files.
func DisableFiles() {
	mutex.Lock()
	defer mutex.Unlock()

	if !logOutputEnabled {
		return
	}

	for _, log := range logs {
		// we'll let the kernel optimize, just do it all parallel
		go log.disableFileOutput()
	}

	logOutputEnabled = false
}

func (l *Log) enableFileOutput() {
	l.l.Lock()
	defer l.l.Unlock()

	for _, e := range l.pending {
		if err := l.writeEntry(e); err != nil {
			l.emergencyLog(e, err)
		}
	}
	l.writeToFile = true
}

func (l *Log) disableFileOutput() {
	l.l.Lock()
	defer l.l.Unlock()

	if l.out != nil {
		l.out.Close()
	}

	l.writeToFile = false
}

func (l *Log) SetConsole(console io.Writer) {
	l.console = console
}

// StreamLines will copy the input line by line as log entries.
func (l *Log) StreamLines(r io.Reader) {
	in := bufio.NewReader(r)
	for {
		line, err := in.ReadBytes('\n')
		if err != nil {
			if err != io.EOF {
				fmt.Fprintf(os.Stderr, "log %s: read lines failed: %v\n", l.name, err)
				time.Sleep(1 * time.Second)
			}
			return
		}
		l.Write(line)
	}
}

// Print to this log.
func (l *Log) Print(v ...interface{}) {
	fmt.Fprint(l, v...)
}

// Printf to this log.
func (l *Log) Printf(pattern string, v ...interface{}) {
	fmt.Fprintf(l, pattern, v...)
}

// Taint is Print to this log with a taint.
func (l *Log) Taint(taint Taint, v ...interface{}) {
	l.append(taint, []byte(fmt.Sprint(v...)))
}

// Taintf is Printf to this log with a taint.
func (l *Log) Taintf(taint Taint, pattern string, v ...interface{}) {
	l.append(taint, []byte(fmt.Sprintf(pattern, v...)))
}

func (l *Log) append(taint Taint, data []byte) {
	// we serialize writes
	l.l.Lock()
	defer l.l.Unlock()

	e := Entry{
		Time:  time.Now(),
		Taint: taint,
		Data:  data,
	}

	console := l.console
	if console != nil {
		buf := &bytes.Buffer{}
		buf.WriteString(string(color.DarkGreen))
		buf.WriteString(e.Time.Format("2006/01/02 15:04:05.000 "))
		buf.WriteString(string(color.Reset))
		buf.WriteString(string(e.Taint.Color()))
		buf.Write(data)
		if data[len(data)-1] != '\n' {
			buf.Write([]byte{'\n'})
		}
		buf.WriteString(string(color.Reset))

		buf.WriteTo(console)
	}

	if !l.writeToFile {
		l.pending = append(l.pending, e)
		// TODO if len(pending) > maxPending { pending = pending[len(pending)-underMaxPending:] }
		// or use a ring
		return
	}

	if err := l.writeEntry(e); err != nil {
		l.emergencyLog(e, err)
	}
}

func (l *Log) emergencyLog(entry Entry, err error) {
	fmt.Fprintf(os.Stderr, "log %s: failed to write entry: %v\n  -> lost entry: ", l.name, err)
	entry.WriteTo(os.Stderr)
}

// Write is part of the io.Writer interface.
func (l *Log) Write(b []byte) (n int, err error) {
	l.append(Normal, b)
	return len(b), nil
}

func (l *Log) writeEntry(e Entry) (err error) {
	ts := e.Time.Truncate(time.Hour).Format(time.RFC3339)

	path := fmt.Sprintf("/var/log/%s.log", l.name)

	if l.outTS != ts {
		if l.out != nil {
			if err := l.out.Close(); err != nil {
				fmt.Fprintf(os.Stderr, "log %s: failed to close output: %v\n", l.name, err)
			}
			archPath := fmt.Sprintf("/var/log/archives/%s.%s.log", l.name, l.outTS)

			os.MkdirAll(filepath.Dir(archPath), 0700)
			if err := os.Rename(path, archPath); err != nil {
				fmt.Fprintf(os.Stderr, "log %s: failed to achive: %v", l.name, err)
			}

			go compress(archPath)
		}
		l.out = nil
		l.outTS = ""
	}

	if l.out == nil {
		l.out, err = os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0640)
		if err != nil {
			return
		}

		l.outTS = ts
	}

	_, err = e.WriteTo(l.out)

	return
}
