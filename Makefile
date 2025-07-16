.PHONY: build build-plugin install-plugin test clean

build:
	go build -o keex ./cmd/keex

build-plugin:
	go build -o kubectl-eex ./cmd/kubectl-eex

install-plugin: build-plugin
	cp kubectl-eex $(GOPATH)/bin/kubectl-eex

test:
	go test ./...

clean:
	rm -f keex kubectl-eex