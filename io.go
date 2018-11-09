package main

import (
	"fmt"
	"sync"
	"time"

	v8 "github.com/ry/v8worker2"
	"golang.org/x/text/encoding/unicode"
)

/* == I/O protocol ==

This is a protocol used by the JavaScript side to request I/O actions
(e.g., write these bytes out), and for the Go side to report results.

The V8Worker2 JavaScript API allows for synchronous send (you call
send with some bytes, and you might get some back) and asynchronous
recv (you call recv with a callback, which may be invoked with some
bytes in the fullness of time).

On the Go side, `SendBytes([]byte) returns void (well, maybe an
error), and you supply a recv callback (`[]byte -> []byte`) when you
create the worker.

There will be different modes of I/O we want to support:

 - a synchronous request vs one to be resolved later (i.e., a Promise)
 - responses of a single value (e.g., a web page) vs a sequence of
   answers (an event stream)

as well as different _kinds_ of request, e.g., for a web page, or for
a local file, or more exotically, for an event stream of some kind,
that we don't know ahead of time.

To be able to dispatch requests, the request includes an identifier
giving the _kind_ of request. The synchronous answer to a request is
either null (assume success, for writes); an error; or, a serial
number indicating a forthcoming reply. It's up to the respective
JavaScript and Go code to agree on what to send and expect for a
particular kind of request.

Because JavaScript characters are UTF-16 (i.e., `String.charCodeAt`
returns 0-65535), data will tend to be aligned on two byte
words. Characters given below are UTF-16 encoded (i.e., take up two
bytes).

```abnf
IO_REQUEST = WRITE_REQUEST / READ_REQUEST

WRITE_MSG = 'W' KIND DATA ; request a write

READ_MSG  = 'R' KIND DATA ; request a read

CHAR = 2*OCTET ; UTF-16 encoded character

KIND = 2*OCTET *CHAR ; length of uint16, then length*2 bytes of UTF-16

DATA = *OCTET

READ_RESPONSE = 'S' SERIAL / ERROR ; introduce sequence ID or error

ERROR = 'E' *CHAR ; error message as string

READ_DATA = 'D' SERIAL DATA / EOF ; continuation of sequence

EOF = 'Z' SERIAL ; no more data, only necessary if a sequence

READ_CANCEL 'C' SERIAL ; stop sending data

SERIAL = 4*OCTET ; uint32 in little-endian order
```

*/

type ReadOrWrite rune

const (
	Read  ReadOrWrite = 'R'
	Write ReadOrWrite = 'W'
)

const trace = false

// ioRecv is a receiver for V8Worker2.Worker that speaks the IO
// protocol detailed above.
type ioRecv struct {
	worker *v8.Worker

	serialMu sync.Mutex
	serial   uint32

	outstandingReqs sync.WaitGroup
}

// handleMsg is a bound method that you can hand to v8worker2.New
func (r *ioRecv) handleMsg(req []byte) []byte {
	if trace {
		fmt.Printf("[TRACE] %#v\n", req)
	}
	var bigEndian bool
	op := req[0]
	// The first 16 bits are an ASCII character, so 0x00nn. On a
	// big-endian computer, the first byte will be 0; on a
	// little-endian computer, the second byte will be 0.
	if op == 0 {
		bigEndian = true
		op = req[1]
	}
	switch op {
	case 'W':
		_, offset := decodeKind(req[2:], bigEndian) // ignore the kind for now, treat as print
		data := decodeStringData(req[offset+2:], bigEndian)
		print(data)
	case 'R':
		kind, offset := decodeKind(req[2:], bigEndian)
		data := req[offset+2:]
		r.serialMu.Lock()
		serial := r.serial
		r.serial += 1
		r.serialMu.Unlock()
		r.outstandingReqs.Add(1)
		repeats := 1
		if trace {
			fmt.Printf("[TRACE] read: %q\n", kind)
		}
		if kind == "repeated" {
			repeats = 5
		}
		go func() {
			for i := 0; i < repeats; i++ {
				time.Sleep(time.Second)
				if err := r.worker.SendBytes(encodeData(serial, []byte(data), bigEndian)); err != nil {
					println("err:", err.Error()) // TODO panic?
				}
			}
			r.outstandingReqs.Done()
		}()
		return encodeResponse('S', serial, bigEndian)
	default:
		panic("got something other than 'W' or 'R'")
	}
	return nil
}

func (r *ioRecv) waitForIO() {
	r.outstandingReqs.Wait()
}

// decodeKind decodes the length-prefixed string representing the kind
// of read or write requested.
func decodeKind(buf []byte, bigEndian bool) (string, int) {
	var len int
	if bigEndian {
		len = int(buf[0]<<8) + int(buf[1])
	} else {
		len = int(buf[0]) + int(buf[1]<<8)
	}
	offset := len*2 + 2
	return decodeStringData(buf[2:offset], bigEndian), offset
}

func decodeStringData(buf []byte, bigEndian bool) string {
	decoder := unicode.UTF16(unicode.Endianness(!bigEndian), unicode.IgnoreBOM).NewDecoder()
	d, err := decoder.Bytes(buf)
	if err != nil {
		panic(err)
	}
	return string(d)
}

func encodeResponse(op rune, serial uint32, bigEndian bool) []byte {
	var b1, b2, b3, b4 byte = byte(serial >> 24), byte(serial >> 16), byte(serial >> 8), byte(serial)
	if bigEndian {
		return []byte{0, byte(op & 0xff), b1, b2, b3, b4}
	} else {
		return []byte{byte(op & 0xff), 0, b4, b3, b2, b1}
	}
}

func encodeData(serial uint32, data []byte, bigEndian bool) []byte {
	var b1, b2, b3, b4 byte = byte(serial >> 24), byte(serial >> 16), byte(serial >> 8), byte(serial)
	if bigEndian {
		return append([]byte{0, 'D', b1, b2, b3, b4}, data...)
	} else {
		return append([]byte{'D', 0, b4, b3, b2, b1}, data...)
	}
}
