#!/bin/sh

# Arguments:
# $1: `slug` - the exercise slug, e.g. `two-fer` (currently ignored)
# $2: `input_dir` - the path containing the solution to test (without trailing slash or preceeding ./)
# $3: `output_dir` - the output path for the test results path (without trailing slash or preceeding ./)

# Example:
# ./run.sh two-fer path/to/two-fer path/to/output/directory
# ./run.sh two-fer twofer outdir

export GOPATH=/go
export PATH="$GOPATH/bin:/usr/local/go/bin:$PATH"

cd "$2" || exit
/opt/test-runner/test-runner $2 $3
