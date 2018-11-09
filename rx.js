import { registerIOCallbacks, sendStringRequest, decodeString } from 'io.js'
import { Subject } from 'node_modules/rxjs/_esm2015/index.js'
import { map } from 'node_modules/rxjs/_esm2015/operators/index.js'

// requestObservable returns an Observable given a data request.
function resultAsObservable(result) {
  const stream = new Subject()
  registerIOCallbacks(result, stream.next.bind(stream), stream.error.bind(stream))
  return stream.asObservable();
}

// TODO better name!
function stringObservable(something) {
  const s = resultAsObservable(sendStringRequest('R', 'repeated', something));
  return s.pipe(map(bytes => decodeString(new Uint16Array(bytes), 0)));
}

// Considerations for design of this module:
// (reference https://rxjs-dev.firebaseapp.com/guide/subject)
//
// The idiomatic use of Observable is that you generate the values
// _per_ subscriber. So it's more like the Observable represents a
// topic, rather than a request.
//
// A Subject is a multicaster, so it fits the I/O purpose a bit better
// -- you might want to watch something upstream and feed several
// calculations from it. So it might be better to just have functions
// which return Subjects.

export { stringObservable };
