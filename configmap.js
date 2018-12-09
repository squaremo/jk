import std from 'std';

class ConfigMap {
  constructor(name, data) {
    this.apiVersion = 'v1';
    this.kind = 'ConfigMap';
    this.meta = {
      name: name,
    };
    this.data = data;
  };
}

function readAsString(f) {
  return std.read(f).then(bytes => String.fromCharCode(...bytes));
}

async function dataFromDir(path) {
  const dir = std.dir(path);
  const files = dir.files.filter(f => !f.isdir);
  const contents = await Promise.all(files.map(f => readAsString(`${path}/${f.name}`)));
  const pairs = files.map((f, i) => [f.name, String.fromCharCode(...contents[i])]);
  return pairs.reduce((acc, [name, content]) => {
    acc[name] = content;
    return acc
  }, {});
}

dataFromDir('examples/config').then(
  data => std.write(new ConfigMap('config', data), '', { format: std.Format.YAML }),
  error => V8Worker2.print(error)
);
