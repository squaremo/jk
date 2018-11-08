import { sendStringRequest } from 'io.js';

function writeJSON(value) {
  const json = JSON.stringify(value);
  return writeString(json);
}

function writeString(value) {
  const answer = sendStringRequest('W', 'stdout', value);
  if (answer !== undefined) {
    throw new Error("did not get null from async write: "+answer.toString());
  }
}

export { writeJSON, writeString };
