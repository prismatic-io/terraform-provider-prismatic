.PHONY: init check format testacc docs build install

init:
	go get
	go mod verify

check:
	test -z "`gofmt -l -s .`"

format:
	go mod tidy
	gofmt -l -w -s .

testacc:
	TF_ACC=1 go test ./... -v $(TESTARGS) -timeout 120m

docs:
	go generate

build:
	go build -v .

install:
	go install
