RUN_CONTEXT ?= docker-compose run --rm akapun

bash:
	docker-compose run --rm akapun /bin/bash

build:
	docker-compose build

go/run:
	$(RUN_CONTEXT) go run main.go

go/test:
	$(RUN_CONTEXT) go test .

go/lint:
	$(RUN_CONTEXT) golangci-lint run .

go/tidy:
	$(RUN_CONTEXT) go mod tidy

go/build:
	$(RUN_CONTEXT) GOOS=linux go build main.go

zip: go/build
	zip function.zip main
