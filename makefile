GOFMT_FILES?=$$(find . -name '*.go')

fmt:
	gofmt -w $(GOFMT_FILES)

build:
	rm -rf bin/ter*
	go fmt ./...
	go build -o ./terraform_test
	tar czvf ./terraform_test.tgz ./terraform_test
	mv ./terraform_test.tgz ./bin/terraform_test.tgz
	rm -rf ./terraform_test

linux:
	rm -rf bin/
	go fmt ./...
	GOOS=linux GOARCH=amd64  go build -o ./terraform_test
	tar czvf terraform_test_linux.tgz ./terraform_test
	rm -rf ./terraform_test