FROM golang:1.27.3-alpine3.24@sha256:d200d2dd76b03a99b52651fd1325aff7dcd35422c35c01bac56571b0507f3deb

# Add a non-root user to run our code as
RUN adduser --disabled-password appuser

# Copy the source code into the container
# and make sure appuser owns all of it
COPY --chown=appuser:appuser . /opt/test-runner

# Build and run the testrunner with appuser
USER appuser

# Default is 'go telemetry local' which saves telemetry locally.
# Since data will never be uploaded, turn it off to avoid unnecessary file writes.
RUN go telemetry off

# Fill the build cache, so each run only has to compile and vet the student's code:
# - Build the standard library and command packages.
# - Build the external packages.
# - Vet every standard library package once, the way go test does (see bin/warm-vet-cache.sh).
# - Then build the test runner.
#
# Go updates a cache file's modification time each time it uses the file, if that time is
# more than an hour old. In a container, changing a file from the image makes Docker copy
# the whole file into the container first. Without the step below, once the image is an
# hour old, every run will copy about a thousand cache files, which takes about 2 seconds.
# Setting the times to the year 2100 stops Go from updating them. This has to happen in
# the same step that creates the files, or Docker will store a second copy of the cache.
RUN go build std cmd \
 && cd /opt/test-runner/external-packages && go mod download && go build ./... \
 && /opt/test-runner/bin/warm-vet-cache.sh \
 && find "$(go env GOCACHE)" -type f -exec touch -d "2100-01-01 00:00:00" {} + \
 && cd /opt/test-runner && go build -o /opt/test-runner/bin/test-runner /opt/test-runner

WORKDIR /opt/test-runner
ENTRYPOINT ["sh", "/opt/test-runner/bin/run.sh"]
