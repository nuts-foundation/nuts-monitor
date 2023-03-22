.PHONY: test

test:
	go test ./...

frontend:
	npm install

docker:
	docker build -t nutsfoundation/nuts-monitor:latest .
