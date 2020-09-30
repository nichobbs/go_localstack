#!/bin/bash

GOLANGCI_LINT_VER="1.31.0"

echo "Formatting project!"
gofmt -s -w .

echo "Adjusting imports!"
dirs=$(go list -f {{.Dir}} ./...)
for d in $dirs; do goimports -w $d/*.go; done

echo "Building project!"
go build ./...

if [ -e bin/golangci-lint ] && [ $(bin/golangci-lint --version | cut -d " " -f4) != "$GOLANGCI_LINT_VER" ]; then
	echo "Removing old golangci-lint!"
	rm bin/golangci-lint
fi

if [ ! -e bin/golangci-lint ]; then
  echo "Installing golangci-lint!"
  curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s "v$GOLANGCI_LINT_VER"
fi

echo "Linting project!"
./bin/golangci-lint run -c .golangci.yml
