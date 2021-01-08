FROM golang:1.15-buster

# add a non-root user to run our code as
RUN adduser --disabled-password --gecos "" appuser

WORKDIR /opt/test-runner
COPY . .

# Build the test runner
RUN GOOS=linux GOARCH=amd64 go build -o /opt/test-runner/bin/test-runner /opt/test-runner

USER appuser
ENV GOCACHE=/tmp

ENTRYPOINT ["sh", "/opt/test-runner/bin/run.sh"]
