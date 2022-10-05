FROM golang:1.18-alpine3.15

# add addtional packages needed for the race detector to work
RUN apk add --update build-base make 

# add a non-root user to run our code as
RUN adduser --disabled-password --gecos "" appuser

WORKDIR /opt/test-runner
COPY . .

# Install extra packages
WORKDIR /opt/test-runner/extra-packages
RUN ./install_dependencies.sh

# Build the test runner
WORKDIR /opt/test-runner
RUN GOOS=linux GOARCH=amd64 go build -o /opt/test-runner/bin/test-runner /opt/test-runner

USER appuser
ENV GOCACHE=/tmp

ENTRYPOINT ["sh", "/opt/test-runner/bin/run.sh"]
