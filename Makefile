.PHONY: build build-plugin install-plugin test cover cover-html clean

build:
	go build -o keex ./cmd/keex

build-plugin:
	go build -o kubectl-eex ./cmd/kubectl-eex

install-plugin: build-plugin
	cp kubectl-eex $(GOPATH)/bin/kubectl-eex

test:
	go test ./...

cover:
	go test -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -func=coverage.out

cover-html: cover
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

clean:
	rm -f keex kubectl-eex coverage.out coverage.html