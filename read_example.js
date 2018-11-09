import { readString } from 'read.js';
import { writeString } from 'write.js'

writeString("asking ...\n")
readString("hello").then(s => writeString(s + '\n'));
