.PHONY: build build-plugin install-plugin test clean

build:
	go build -o kenv ./cmd/kenv

build-plugin:
	go build -o kubectl-eextract ./cmd/kubectl-eextract

install-plugin: build-plugin
	cp kubectl-eextract $(GOPATH)/bin/kubectl-eextract

test:
	go test ./...

clean:
	rm -f kenv kubectl-eextract