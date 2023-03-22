.PHONY: test

test: backend-test feature-test

feature-test:
	NODE_ENV=test npm install
	NODE_ENV=test npm run build
	NODE_ENV=test npm run test

backend-test:
	go test ./... -race

docker:
	docker build -t nutsfoundation/nuts-monitor:latest .
