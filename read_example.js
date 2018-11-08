import { readString } from 'read.js';

V8Worker2.print("asking ...")
readString("unimportant").then(V8Worker2.print);
