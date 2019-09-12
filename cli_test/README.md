# CLI Integration tests

Run the full suite:

```bash
make build
go test -v -p 4 ./cli_test -tags cli_test
```

`-v` for verbose, `-p 4` to use 4 cores, `-tags cli_test` a build tag (specified in `cli_test.go`) to tell go not to ignore the package

> NOTE: While the full suite runs in parallel, some of the tests can take up to a minute to complete

> NOTE: The tests will use the `kvd` or `kvcli` binaries in the build dir. Or in `$BUILDDIR` if that env var is set.