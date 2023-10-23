FROM golang:1.21.1-alpine3.17

# Add addtional packages needed for the race detector to work
RUN apk add --update build-base make

# Add a non-root user to run our code as
RUN adduser --disabled-password appuser

# Copy the source code into the container
# and make sure appuser owns all of it
COPY --chown=appuser:appuser . /opt/test-runner

# Build and run the testrunner with appuser
USER appuser

# This populates the build cache with the standard library
# and command packages so future compilations are faster
RUN go build std cmd

# Install external packages
WORKDIR /opt/test-runner/external-packages
RUN go mod download
# Populate the build cache with the external packages
RUN go build ./...

# Build the test runner
WORKDIR /opt/test-runner
RUN go build -o /opt/test-runner/bin/test-runner /opt/test-runner

ENTRYPOINT ["sh", "/opt/test-runner/bin/run.sh"]
