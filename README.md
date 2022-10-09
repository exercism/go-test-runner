# Exercism's Go Test Runner

![Go CI](https://github.com/exercism/go-test-runner/workflows/Run%20linter%20and%20tests/badge.svg)
[![Coverage Status](https://coveralls.io/repos/github/exercism/go-test-runner/badge.svg)](https://coveralls.io/github/exercism/go-test-runner)

This is [Exercism's test runner](https://github.com/exercism/v3-docs/tree/master/anatomy/track-tooling/test-runners#test-runners) for the Go track.

## Executing the Test Runner

The test runner requires 2 parameters:
- `input_dir`: the path containing the solution to test
- `output_dir`: the output path for the test results

### Local Development

```bash
go run . testrunner/testdata/practice/passing outdir
```

#### Run the package tests

```bash
go test ./...
```

#### Run the linter

Linting (and testing) is performed in a [github action workflow - test.yml](.github/workflows/test.ym). You can [install golangci-lint locally](https://golangci-lint.run/usage/install/#local-installation) and then run:

```bash
golangci-lint run ./...
```

#### Interactive Debug / REPL

The original AST parsing code was developed [using a Jupyter interactive Go REPL](https://jupyter.readthedocs.io/en/latest/install/notebook-classic.html) thanks to the [gophernotes project](https://github.com/gopherdata/gophernotes). Consult the gophernotes docs for installation instructions. Once installed, you should be able to view, run, and modify the provided debug code "live" without constantly recompiling:

```bash
# assuming python3 with notebook installed via pip3, ymmv
python3 -m notebook ast_debug.ipynb
```

### Docker

A docker container is used to run the test runner against submitted exercises. To build the container locally, execute the following from the repository `root` directory:

```bash
docker build -t exercism/go-test-runner .
```

Run the test runner in the container by passing in the slug name, and absolute paths to the exercise (solution) and a writeable tmp directory. These directories should be mounted as volumes:

```bash
docker run --network none --read-only -v $(pwd)/testrunner/testdata/practice/gigasecond:/solution -v /tmp:/tmp exercism/go-test-runner gigasecond /solution /tmp
```

For troubleshooting / debug you can name the container, run it in interactive mode, and detach from it using:

```bash
docker run --name exercism-go-test-runner -d -i --network none --read-only -v $(pwd)/testrunner/testdata/practice/gigasecond:/solution -v /tmp:/tmp exercism/go-test-runner gigasecond /solution /tmp
# You can then access the container as follows:
docker exec -it --user appuser $(docker ps -q --filter name=exercism-go-test-runner) /bin/sh
```

### Extra Go packages

Some extra Go packages are downloaded when the docker image is built.
This allows students to use these extra packages in their solutions.

The list of extra packages and their versions is in `extra-packages/go.mod`.

To add or remove a package from the list of extra packages supported:

1. Add/remove the corresponding import from `extra-packages/deps.go`
2. Run `go mod tidy` inside the `extra-packages` directory
3. Commit `deps.go` along with the changes to `go.mod` and `go.sum` produced by `go mod tidy`.

## Subtests

The test runner is responsible for [returning the `test_code` field](https://github.com/exercism/v3-docs/blob/master/anatomy/track-tooling/test-runners/interface.md#command), which should be a copy of the test code corresponding to each test result. 

For top-level tests, the AST is used to return the function code directly. For [tests containing subtests](https://blog.golang.org/subtests), additional processing is required. To ease the burden of advanced AST processing on unstructured / non deterministic test code, subtests should adhere to the following specification. **If a test employs subtests, do not mix it with test or other code outside of the Run() call.** 

- Subtests not meeting the spec will be treated as top-level tests, with the entire test function code being returned for every subtest.
- Assertions/outputs made outside of the Run() call will not be included in the result JSON because the "parent" tests are removed from the results if subtests are present. (Parent test reports were confusing to students because they did not include any assertion or `fmt.Println` output.)

At some point, we may [implement a static analyzer](https://rauljordan.com/2020/11/01/custom-static-analysis-in-go-part-1.html) which warns the exercise submitter when they commit subtests not meeting the specification.


### Subtest Format Specification

The specification is annotated in the comments of the following example test:
```go
func TestParseCard(t *testing.T) {
  // There can be additional code here, it will be shown for all subtests.
  // If the code here includes assignments, the test data variable below needs to be called "tests".

  tests := []struct {
    name string // The name field is required
    card string
    want int
  }{
    // The relevant test data will be parsed out individually for each subtest
    {
      // spaces are fine in a test name, but do not mix them with underscores
      // - because the extraction code won't be able to find it
      name: "parse queen",
      card: "queen",
      want: 10,
    },
    // For example, this element will be parsed for `TestParseCard/parse_king`
    {
      name: "parse king",
      card: "king",
      want: 10,
    },
  }

  // There can be additional code here, it will be shown for all subtests.

  // The contents of the function literal will be extracted as the test code
  for _, tt := range tests {
    // The Run() call must be the first statement in the for loop
    t.Run(tt.name, func(t *testing.T) {
      // This code block will be pulled into the resulting test_code field
      if got := ParseCard(tt.card); got != tt.want {
        t.Errorf("ParseCard(%s) = %d, want %d", tt.card, got, tt.want)
      }
    })
  }

  // There can be additional code here, it will be shown for all subtests.
}
```

The test code above will result in the following `test_code` field, corresponding to the test named `TestParseCard/parse_queen`:
```go
tt := struct {
  name string
  card string
  want int
}{
  name: "parse queen",
  card: "queen",
  want: 10,
}
if got := ParseCard(tt.card); got != tt.want {
  t.Errorf("ParseCard(%s) = %d, want %d", tt.card, got, tt.want)
}
```

## Providing Additional Testing Flags

Exercises can supply additional flags that will be included when the test runner executes the `go test` command.
This is done via the `.meta/config.json` file of the exercise. See example below.

```json
{
  // ...
  "custom": {
    "testingFlags": [
      "-race"
    ]
  }
}
```

Currently, only the flag `-race` is supported.
If more flags should be allowed in the future, they first need to be added to the `allowedTestingFlags` list in `testrunner/execute.go`.