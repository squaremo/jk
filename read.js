import { decodeString, sendStringRequest, resultAsPromise } from 'io.js';

// returns a promise which will resolve to the data requested as an
// ArrayBuffer
function read(something) {
  return resultAsPromise(sendStringRequest('R', 'stdin', something));
}

// returns a promise which will resolve to the data requested as a
// string
function readString(something) {
  return resultAsPromise(sendStringRequest('R', 'stdin', something)).then(
    bytes => decodeString(new Uint16Array(bytes), 0)
  );
}

export { read, readString };
