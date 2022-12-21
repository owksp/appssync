NAME = appss
TARGET_LINUX = GOARCH=amd64 GOOS=linux

.PHONY: all
all: test_build test install

.PHONY: test
test:
	go test -covermode=atomic

.PHONY: test_build
test_build:
	go mod verify && go mod tidy
	cd cmd/appssync && go build main.go && rm main

.PHONY: install
install:
	cd cmd/appssync && go build -o ${NAME} main.go && chmod +x ${NAME} && mv ${NAME} ${GOPATH}/bin/${NAME}