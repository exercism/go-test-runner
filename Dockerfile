FROM golang:1.20.2-alpine3.17

# add addtional packages needed for the race detector to work
RUN apk add --update build-base make 

WORKDIR /opt/test-runner
COPY . .

# Install external packages
WORKDIR /opt/test-runner/external-packages
RUN go mod download

# Build the test runner
WORKDIR /opt/test-runner
RUN go build -o /opt/test-runner/bin/test-runner /opt/test-runner

ENTRYPOINT ["sh", "/opt/test-runner/bin/run.sh"]
