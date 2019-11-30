# Exercism's Go Test Runner

This is Exercism's test runner for the Go track.

## Executing the Test Runner

The test runner takes 3 parameters:
- the exercise `slug`, e.g. `two-fer`
- the `path` containing the solution to test
- the `output path` for the test results

Example to execute for development (Go needs to be installed):

```bash
cd path/containing/tests
go test --json . | go run ~/go-test-runner/cmd/testoutput > result.json
```

## Docker

To `build` execute the following from the repositories `root` directory:

```bash
docker build -t exercism/go-test-runner .
```

To `run` from docker pass in the solutions path as a volume and execute with the necessary parameters:

```bash
docker run -v $(PATH_TO_SOLUTION):/solution exercism/go-test-runner ${SLUG} /solution /solution
```

Example:

```bash
docker run -v ~/solution-238382y7sds7fsadfasj23j:/solution exercism/go-test-runner two-fer /solution /solution
```
