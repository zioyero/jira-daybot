all: generate build

generate:
	go generate ./...

build:
	go build -o bin/ ./...

debug: generate
	go build -o bin/ -gcflags "all=-N -l" ./...

run-debug: debug
	./bin/cmd -run-now -output=stdout

run-now: build
	./bin/cmd -run-now -output=slack

dockerize:
	docker build -t zioyero/jira-daybook .

deploy-docker: dockerize
	docker rm -f jira_daybook || true
	docker run --name jira_daybook -d zioyero/jira-daybook
