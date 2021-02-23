TEST?=$$(go list ./... | grep -v 'vendor')
HOSTNAME=hashicorp.com
NAMESPACE=edu
NAME=kuma
BINARY=terraform-provider-${NAME}
VERSION=0.1
OS_ARCH=darwin_amd64

default: install

build:
	go build -o ./bin/${BINARY}

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

install: build
	mkdir -p ~/.terraform.d/plugins/${OS_ARCH}
	mkdir -p ./examples/.terraform/plugins/${OS_ARCH}
	cp ./bin/${BINARY} ~/.terraform.d/plugins/${OS_ARCH}
	cp ./bin/${BINARY} ./examples/.terraform/plugins/${OS_ARCH}

setup:
	curl https://releases.hashicorp.com/terraform/0.12.30/terraform_0.12.30_${OS_ARCH}.zip -o /tmp/terraform_0.12.30.zip
	unzip /tmp/terraform_0.12.30.zip -d ./test

.PHONY: test
test: 
	cd test && PATH=$(PWD)/test:${PATH} go test

testacc: 
	TF_ACC=1 go test $(TEST) -v $(TESTARGS) -timeout 120m