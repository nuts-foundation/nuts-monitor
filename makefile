.PHONY: test

apis:
	oapi-codegen --config codegen/config.yaml api/api.yaml | gofmt > api/generated.go

test: backend-test feature-test

feature-test:
	NODE_ENV=test npm install
	NODE_ENV=test npm run build
	NODE_ENV=test npm run test

backend-test:
	go test ./... -race

docker:
	docker build -t nutsfoundation/nuts-monitor:latest .
