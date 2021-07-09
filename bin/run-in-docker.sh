#!/usr/bin/env bash

# Synopsis:
# Run the test runner on a solution using the test runner Docker image.
# The test runner Docker image is built automatically.

# Arguments:
# $1: exercise slug
# $2: absolute path to solution folder
# $3: absolute path to output directory

# Output:
# Writes the test results to a results.json file in the passed-in output directory.
# The test results are formatted according to the specifications at https://github.com/exercism/docs/blob/main/building/tooling/test-runners/interface.md

# Example:
# ./bin/run-in-docker.sh two-fer /absolute/path/to/two-fer/solution/folder/ /absolute/path/to/output/directory/

# If any required arguments is missing, print the usage and exit
if [ -z "$1" ] || [ -z "$2" ] || [ -z "$3" ]; then
    echo "usage: ./bin/run-in-docker.sh exercise-slug /absolute/path/to/solution/folder/ /absolute/path/to/output/directory/"
    exit 1
fi

slug="$1"
input_dir="${2%/}"
output_dir="${3%/}"

# Create the output directory if it doesn't exist
mkdir -p "${output_dir}"

# Build the Docker image
docker build --rm -t exercism/test-runner .

# Run the Docker image using the settings mimicking the production environment
docker run \
    --rm \
    --network none \
    --read-only \
    --mount type=bind,src="${input_dir}",dst=/solution \
    --mount type=bind,src="${output_dir}",dst=/output \
    --mount type=volume,target=/tmp \
    exercism/test-runner "${slug}" /solution /output
