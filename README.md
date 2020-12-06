# Exercism's Go Test Runner

This is [Exercism's test runner](https://github.com/exercism/v3-docs/tree/master/anatomy/track-tooling/test-runners#test-runners) for the Go track.

## Executing the Test Runner

The test runner takes 2 parameters:
- `input_dir`: the path containing the solution to test
- `output_dir`: the output path for the test results

Example to run for local development (Go needs to be installed):

```bash
go run main.go ~/Exercism/go/gigasecond outdir
```

## Docker

To `build` execute the following from the repository `root` directory:

```bash
docker build -t exercism/go-test-runner .
```

To `run` from docker pass in the solutions path as a volume and execute with the necessary parameters:

```bash
docker run -v $(PATH_TO_SOLUTION):/solution exercism/go-test-runner ${SLUG} /solution /solution
```

Example:

```bash
docker run -v ~/Exercism/go/gigasecond:/solution exercism/go-test-runner gigasecond /solution /solution
```
