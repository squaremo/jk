// writeSerial writes a serial number (uint32) at offset in the given
// array.
function writeSerial(uint16array, serial, offset) {
  uint16array[offset] = serial & 0xffff;
  uint16array[offset+1] = serial >> 16;
  return offset + 2;
}

// readSerial reads a serial number (uint32) from offset in the given
// array.
function readSerial(uint16array, offset) {
  return uint16array[offset] + uint16array[offset+1] << 16;
}

// writeKind writes a length-prefixed string into the array at offset.
function writeKind(uint16array, kind, offset) {
  uint16array[offset] = kind.length;
  writeString(uint16array, kind, offset+1);
  return offset + 1 + kind.length;
}

// readKind reads a length-prefixed string from the array at offset.
function readKind(uint16array, offset) {
  len = uint16array[offset]
  return readString(uint16array.slice(offset+1, offset+1+len), 0)
}

// writeString writes the string into the array at offset.
function writeString(uint16array, string, offset) {
  for (let i = 0, l = string.length; i < l; i++) {
    uint16array[offset+i] = string.charCodeAt(i);
  }
  return offset + string.length;
}

// readString reads a string from offset to the end of an array
function readString(uint16array, offset) {
  return String.fromCharCode(null, ...uint16array.slice(offset))
}

function sendStringRequest(readOrWrite, kind, data) {
  const buf = new ArrayBuffer(4 + kind.length * 2 + data.length * 2);
  const view = new Uint16Array(buf);
  view[0] = readOrWrite.charCodeAt(0);
  var offset = writeKind(view, kind, 1);
  writeString(view, data, offset);
  return decodeResult(V8Worker2.send(buf));
}

function decodeResult(result) {
  if (result !== null) {
    const view = new Uint16Array(result);
    switch (String.fromCharCode(view[0])) {
    case 'S':
      throw new Error('sequences not understood yet');
    case 'E':
      errorStr = readString(view, 1)
      return new Error(errorStr);
    }
  }
  return result
}

export { sendStringRequest };
