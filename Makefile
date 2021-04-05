RUN_CONTEXT ?= docker-compose run --rm akapun

bash:
	docker-compose run --rm akapun /bin/bash

build:
	docker-compose build

go/run:
	$(RUN_CONTEXT) go run main.go

go/test:
	$(RUN_CONTEXT) go test .

go/fmt:
	$(RUN_CONTEXT) go fmt .

go/lint:
	$(RUN_CONTEXT) golangci-lint run .

go/tidy:
	$(RUN_CONTEXT) go mod tidy

go/build:
	GOOS=linux go build main.go

zip: go/build
	zip function.zip main

upload:
	echo "Show a sample command to upload a zip package to Lambda function"
	echo "aws lambda update-function-code --function-name akapun --zip-file fileb://function.zip"
	echo "aws lambda update-function-configuration --function-name akapun --environment Variables={AKASHI_CORP_ID=string\,AKASHI_TOKEN=string}"
