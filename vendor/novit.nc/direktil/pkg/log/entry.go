package log

import (
	"encoding/base32"
	"fmt"
	"io"
	"time"
)

type Entry struct {
	Time  time.Time
	Taint Taint
	Data  []byte
}

// WriteTo writes this log entry to w.
// Automatically appends a new line if it's not already present to
// get an easy to read log.
func (e Entry) WriteTo(w io.Writer) (n int64, err error) {
	l := len(e.Data)
	t := e.Time.UnixNano()

	flags := byte(0)
	appendNL := e.Data[len(e.Data)-1] != '\n'

	if appendNL {
		flags |= AppendNL
	}

	b := []byte{
		flags,
		byte(e.Taint),
		byte(l >> 16 & 0xff),
		byte(l >> 8 & 0xff),
		byte(l >> 0 & 0xff),
		byte(t >> 56 & 0xff),
		byte(t >> 48 & 0xff),
		byte(t >> 40 & 0xff),
		byte(t >> 32 & 0xff),
		byte(t >> 24 & 0xff),
		byte(t >> 16 & 0xff),
		byte(t >> 8 & 0xff),
		byte(t >> 0 & 0xff),
	}

	// the binary part is b32 encoded. Obscure but still readable in text mode.
	enc := base32.StdEncoding

	headerLen := enc.EncodedLen(len(b))
	baLen := headerLen + len(e.Data)
	if appendNL {
		baLen++
	}
	ba := make([]byte, baLen)
	enc.Encode(ba, b)

	copy(ba[headerLen:], e.Data)

	if appendNL {
		ba[baLen-1] = '\n'
	}

	nw, err := w.Write(ba)
	return int64(nw), err
}

// ReadFrom reads the next entry from r, updating this entry.
func (e *Entry) ReadFrom(r io.Reader) (n int64, err error) {
	enc := base32.StdEncoding

	const L = 1 + 1 + 3 + 8
	b := make([]byte, L)
	ba := make([]byte, enc.EncodedLen(L))

	nr, err := r.Read(ba)
	if err != nil {
		return int64(nr), err
	}
	fmt.Println(string(ba))

	enc.Decode(b, ba)
	fmt.Println(b)

	p := 0
	flags := b[p]
	p++

	e.Taint = Taint(b[p])
	p++

	l := int32(0)
	for i := 0; i < 3; i++ {
		l = l<<8 | int32(b[p])
		p++
	}

	readLen := l
	if flags|AppendNL != 0 {
		readLen++
	}

	t := int64(0)
	for i := 0; i < 8; i++ {
		t = t<<8 | int64(b[p])
		p++
	}
	e.Time = time.Unix(0, t)

	data := make([]byte, readLen)
	m, err := r.Read(data)
	n += int64(m)
	if err != nil {
		return n, err
	}

	e.Data = data[:l]

	return n, nil
}
