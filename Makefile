BIN := mackerel-plugin-fireworq
GOBIN ?= $(shell go env GOPATH)/bin
export GO111MODULE=on

.PHONY: all
all: clean build

.PHONY: build
build:
	go build -o $(BIN) .

.PHONY: install
install:
	go install ./...

.PHONY: cross
cross: $(GOBIN)/goxz
	goxz -n $(BIN) .

$(GOBIN)/goxz:
	GO111MODULE=off go get github.com/Songmu/goxz/cmd/goxz

.PHONY: test
test: build
	go test -v ./...

.PHONY: lint
lint: $(GOBIN)/golint
	go vet ./...
	golint -set_exit_status ./...

$(GOBIN)/golint:
	GO111MODULE=off go get golang.org/x/lint/golint

.PHONY: clean
clean:
	rm -rf goxz
	go clean

.PHONY: bump
bump: $(GOBIN)/gobump
	git tag "v$(shell sh -c 'read -p "input next version: " v && echo $$v')"
	git push
	git push --tags

.PHONY: upload
upload: $(GOBIN)/ghr
	ghr "v$(VERSION)" goxz

$(GOBIN)/ghr:
	GO111MODULE=off go get github.com/tcnksm/ghr

.PHONY: release
release: test lint clean bump cross upload
