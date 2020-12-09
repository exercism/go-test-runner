# Exercism's Go Test Runner

This is [Exercism's test runner](https://github.com/exercism/v3-docs/tree/master/anatomy/track-tooling/test-runners#test-runners) for the Go track.

## Executing the Test Runner

The test runner requires 2 parameters:
- `input_dir`: the path containing the solution to test
- `output_dir`: the output path for the test results

### Local Development

```bash
go run ./... ~/Exercism/go/gigasecond outdir
```

### Docker

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

## Subtests

The test runner is responsible for [returning the `test_code` field](https://github.com/exercism/v3-docs/blob/master/anatomy/track-tooling/test-runners/interface.md#command), which should be a copy of the test code corresponding to each test result. 

For regular tests, the AST is used to return the code directly. For [tests containing subtests](https://blog.golang.org/subtests), additional processing is required. To ease the burden of advanced AST processing on unstructured / non deterministic test code, subtests should adhere to the following specification. Subtests not meeting the spec will be treated as regular tests, with the entire test function code being returned for every subtest. 


### Subtest Format Specification

```go
func TestParseCard(t *testing.T) {
  // The table data must be created first, and must be named `tests`
  tests := []struct {
    name string // The name field is required
    card string
    want int
  }{
    // The relevant test data will be parsed out individually for each subtest
    {
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

  // The test loop must follow immediately after the table definition
  // The element should be named `tt`
  // The contents of the function literal will be extracted as the test code
  for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
      if got := ParseCard(tt.card); got != tt.want {
        t.Errorf("ParseCard(%s) = %d, want %d", tt.card, got, tt.want)
      }
    })
  }
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
t.Run(tt.name, func(t *testing.T) {
  if got := ParseCard(tt.card); got != tt.want {
    t.Errorf("ParseCard(%s) = %d, want %d", tt.card, got, tt.want)
  }
})
```
