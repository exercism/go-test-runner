#!/bin/sh
# Runs `go test` once on a package that imports every public standard library
# package. `go test` vets each imported package before it compiles the tests,
# using its own set of checks. Doing that here stores the results in the build
# cache, so a student run only vets the student code.
set -e
dir=$(mktemp -d)
cd "$dir"
go mod init warm >/dev/null 2>&1
{
  echo "package warm"
  echo "import ("
  go list std | grep -v -E "(^|/)internal(/|$)|^cmd/|^vendor/|^runtime/cgo$|^plugin$|^syscall/js$|^crypto/internal" | sed "s/.*/\t_ \"&\"/"
  echo ")"
} > warm.go
echo "package warm" > warm_test.go
go test -run "^$" . >/dev/null
cd /
rm -rf "$dir"
