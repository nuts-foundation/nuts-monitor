.PHONY: test

test:
	go test ./...

docker:
	docker build -t nutsfoundation/nuts-monitor:latest .
