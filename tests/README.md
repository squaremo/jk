# Integration tests

## Running integration tests

```console
$ go test -v ./tests
```

## Adding a new test

To add a test, drop a `test-$testname.js` file in this directory. It will be
automatically picked up and run using `jk` and its results will be compared
to the results we expect.

- The test will run with:

```console
$ jk run -o test-$testname.expected test-$testname.js
```

- If the file `test-$testname.js.skip` exists, the test is skipped. This is
  useful to commit failing tests but not make them part of the test suite.

- If the file `test-$testname.js.error` exists, we'll check that `jk` exits
  with an error. Otherwise, we expect `jk` to exit with 0.

- If the file `test-$testname.js.cmd` exists, it's content is used as the
  command line to execute jk. This allows to:

    1. Use custom jk commands or options.
    2. Run js files that aren't in the `tests/` directory.

  `.cmd` files look like:

  ```text
  jk run %t/test.js
  ```

  In that file, special variables can be used for convenience:

  **%t**: the name of the test

  **%d**: the name of the recommended output directory (ie. `test-%testname.got`)


- `jk` std{out,err} will be compared to `test-$testname.js.expected`.

- If `jk` writes files to disk, they will be compared to the files in the
  `test-$testname.expected` directory.
