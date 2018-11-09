import { stringObservable } from 'rx.js'
import { writeString } from 'write.js'

stringObservable('whatever').forEach(s => writeString(s + '\n'));
