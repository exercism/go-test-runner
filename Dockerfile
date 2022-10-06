FROM golang:1.18-alpine3.15

# add addtional packages needed for the race detector to work
RUN apk add --update build-base make 

# add a non-root user to run our code as
RUN adduser --disabled-password --gecos "" appuser

WORKDIR /opt/test-runner
COPY . .

# Download all dependencies.
# We want some more than what is needed in the build below.
RUN go mod download

# Build the test runner
RUN GOOS=linux GOARCH=amd64 go build -o /opt/test-runner/bin/test-runner /opt/test-runner

USER appuser
ENV GOCACHE=/tmp

ENTRYPOINT ["sh", "/opt/test-runner/bin/run.sh"]
