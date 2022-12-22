CLI = simpleas
#TARGET_LINUX = GOARCH=amd64 GOOS=linux

.PHONY: all
all: test_build test install

.PHONY: test
test:
	go test -covermode=atomic

.PHONY: test_build
test_build:
	go mod verify && go mod tidy
	cd cmd/cli && go build main.go && rm main

.PHONY: install
install:
	cd cmd/cli && go build -o ${CLI} main.go && chmod +x ${CLI} && mv ${CLI} ${GOPATH}/bin/${CLI}