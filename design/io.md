# Design for doing I/O

 1. Every read returns a `Promise`
 1. Use URLs to specify requests
 1. Handlers of whatever kind get registered per scheme
 1. Other RPCs can return a promise too
 1. Filesystem info (listing dirs, stat) are RPCs not reads

## Rationale

 - use URLs to specify what is requested
   - paths without a scheme can be a shortcut for file://

 - even local file reads don't need to block JS execution
   --> - make all reads return a `Promise`

 - some RPCs will want to have promise results as well, e.g., timer
   --> - use protocol for returning ticket, and result later
       - one RPC is read(URL)

 - some RPCs will want to have repeated callbacks, e.g., ticker
   --> - ticket/response protocol should distinguish between one-off results and stream entries (and end of stream)
       - requests should be cancellable

 - up to JS to decide if it got the right response

## Protocol

```
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

## Handling repeated actions
