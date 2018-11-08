import k from 'kubernetes.js';
import { writeJSON } from 'write.js';

const container = k.Container('nginx', 'nginx:1.15.4');
const deployment = k.Deployment('nginx', 3, [container]);
writeJSON(deployment);
