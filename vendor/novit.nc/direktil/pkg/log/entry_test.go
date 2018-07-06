package log

import (
	"bytes"
	"testing"
	"time"
)

func TestEncode(t *testing.T) {
	l1 := Entry{time.Now(), Normal, []byte("test entry")}
	l2 := Entry{}

	buf := &bytes.Buffer{}
	l1.WriteTo(buf)

	t.Log(buf.String())

	_, err := l2.ReadFrom(buf)
	if err != nil {
		t.Error("read error: ", err)
	}

	if l1.Taint != l2.Taint {
		t.Errorf("wrong taint: %v != %v", l1.Taint, l2.Taint)
	}

	if l1.Time.UnixNano() != l2.Time.UnixNano() {
		t.Errorf("wrong time: %v != %v", l1.Time, l2.Time)
	}

	if !bytes.Equal(l1.Data, l2.Data) {
		t.Errorf("wrong data: %q != %q", string(l1.Data), string(l2.Data))
	}

	if l := len(buf.Bytes()); l > 0 {
		t.Errorf("%d bytes not read", l)
	}
}
