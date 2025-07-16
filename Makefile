.PHONY: build build-plugin install-plugin test clean

build:
	go build -o kenv ./cmd/kenv

build-plugin:
	go build -o kubectl-kenv ./cmd/kubectl-kenv

install-plugin: build-plugin
	cp kubectl-kenv $(GOPATH)/bin/kubectl-kenv

test:
	go test ./...

clean:
	rm -f kenv kubectl-kenv