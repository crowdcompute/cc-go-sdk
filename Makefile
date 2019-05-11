DIR := ${CURDIR}
BINARY_NAME		=	goclient
VERSION         :=	$(shell cat ./VERSION)

build:
	go build -ldflags "-X main.Version=$(VERSION)" -o "./build/bin/$(BINARY_NAME)" main.go
