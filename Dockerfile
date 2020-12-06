FROM golang:1.15-alpine as builder

# Install SSL ca certificates
RUN apk update && apk add git && apk add ca-certificates

# Create appuser
RUN adduser -D -g '' appuser && \
    mkdir /src

# The working directory should be /opt/test-runner
WORKDIR /src
COPY ./go.mod /src/go.mod

# download dependencies
RUN go mod download

# Copy run.sh to /go/bin which we later copy to the final container
COPY ./run.sh /go/bin/bin/run.sh
RUN chmod +x /go/bin/bin/run.sh

# get the rest of the source code
COPY . /src

# build
RUN GOOS=linux GOARCH=amd64 go build -o /go/bin/test-runner /src/

# Build a minimal and secured container
# To run the tests we need Go installed.
# Therefore, unfortunately we cannot build from scratch as we would normally do with Go.
FROM golang:1.15-alpine
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /go/bin/ /opt/test-runner

RUN mkdir /home/appuser/.cache/go-build -p && \
    chown -R appuser /home/appuser

USER appuser
WORKDIR /opt/test-runner

ENTRYPOINT ["/opt/test-runner/bin/run.sh"]
