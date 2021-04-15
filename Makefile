NAME=kuma
BINARY=terraform-provider-${NAME}
VERSION=0.1
OS_ARCH=$(shell uname -s | awk '{print tolower($$0)}' | sed "s/darwin/osx/")-amd64

.PHONY: lint
lint:
	golangci-lint run ./...

.PHONY: check
check: lint

.PHONY: default
default: install

.PHONY: build
build:
	go build -o ./bin/${BINARY}

.PHONY: release
release:
	GOOS=darwin GOARCH=amd64 go build -o ./bin/${BINARY}_${VERSION}_darwin_amd64
	GOOS=freebsd GOARCH=386 go build -o ./bin/${BINARY}_${VERSION}_freebsd_386
	GOOS=freebsd GOARCH=amd64 go build -o ./bin/${BINARY}_${VERSION}_freebsd_amd64
	GOOS=freebsd GOARCH=arm go build -o ./bin/${BINARY}_${VERSION}_freebsd_arm
	GOOS=linux GOARCH=386 go build -o ./bin/${BINARY}_${VERSION}_linux_386
	GOOS=linux GOARCH=amd64 go build -o ./bin/${BINARY}_${VERSION}_linux_amd64
	GOOS=linux GOARCH=arm go build -o ./bin/${BINARY}_${VERSION}_linux_arm
	GOOS=openbsd GOARCH=386 go build -o ./bin/${BINARY}_${VERSION}_openbsd_386
	GOOS=openbsd GOARCH=amd64 go build -o ./bin/${BINARY}_${VERSION}_openbsd_amd64
	GOOS=solaris GOARCH=amd64 go build -o ./bin/${BINARY}_${VERSION}_solaris_amd64
	GOOS=windows GOARCH=386 go build -o ./bin/${BINARY}_${VERSION}_windows_386
	GOOS=windows GOARCH=amd64 go build -o ./bin/${BINARY}_${VERSION}_windows_amd64

.PHONY: install
install: build
	@mkdir -p ./examples/.terraform/plugins/${OS_ARCH}
	@cp ./bin/${BINARY} ./examples/.terraform/plugins/${OS_ARCH}
	@echo "Install provider into ./examples/.terraform/plugins/${OS_ARCH}"

.PHONY: setup
setup:
	curl https://releases.hashicorp.com/terraform/0.12.30/terraform_0.12.30_${OS_ARCH}.zip -o /tmp/terraform_0.12.30.zip
	unzip /tmp/terraform_0.12.30.zip -d ./test

.PHONY: test
test: 
	cd test && PATH=$(PWD)/test:${PATH} go test
