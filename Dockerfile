FROM golang:1.27.1-alpine3.24@sha256:8a5910f31396cd4d89662f56c68b3ae31d374308270a1c3bd96672ee5ed43414

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

# This populates the build cache with the standard library
# and command packages so future compilations are faster
RUN go build std cmd

# Populate the build cache with the external packages
RUN cd /opt/test-runner/external-packages && go mod download && go build ./...

# Build the test runner
RUN cd /opt/test-runner && go build -o /opt/test-runner/bin/test-runner /opt/test-runner

WORKDIR /opt/test-runner
ENTRYPOINT ["sh", "/opt/test-runner/bin/run.sh"]
