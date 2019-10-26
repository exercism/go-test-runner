FROM golang:1.12-alpine as builder

# Install SSL ca certificates
RUN apk update && apk add git && apk add ca-certificates

# The working directory should be /opt/test-runner
WORKDIR /opt/test-runner
RUN mkdir /opt/test-runner/output

# Create appuser
RUN adduser -D -g '' appuser
RUN mkdir /home/appuser/.cache/go-build -p


# source code
COPY ./cmd/testoutput/main.go /opt/test-runner/main.go
COPY ./go.mod /opt/test-runner/go.mod
COPY ./go.sum /opt/test-runner/go.sum

COPY ./exreport /opt/test-runner/exreport

# download dependencies
RUN go mod download

# Create run.sh
COPY ./run.sh /opt/test-runner/bin/run.sh
RUN chmod +x /opt/test-runner/bin/run.sh

# set file permissions
RUN chown -R appuser /home/appuser/
RUN chown -R appuser /opt/test-runner
RUN chown -R appuser /home/appuser/
RUN chown -R appuser /go/

# build
RUN go build /opt/test-runner/main.go
RUN GOOS=linux GOARCH=amd64 go build --tags=build -o /go/bin/test-runner .

# Build a minimal and secured container
# The ast parser needs Go installed for import statements.
# Therefore, unfortunately we cannot build from scratch as we would normally do with Go.
FROM golang:1.12-alpine
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /go/ /go/
COPY --from=builder /opt/test-runner/ /opt/test-runner
COPY --from=builder /home/appuser/ /home/appuser/

USER appuser
WORKDIR /opt/test-runner

ENTRYPOINT ["/opt/test-runner/bin/run.sh"]