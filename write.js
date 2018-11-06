import { sendStringRequest } from './io.js';

function write(value) {
  const json = JSON.stringify(value);
  const answer = sendStringRequest('W', 'stdout', json);
  if (answer !== undefined) {
    throw new Error("did not get null from async write: "+answer.toString());
  }
}

export default write;
