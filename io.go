package main

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



*/
