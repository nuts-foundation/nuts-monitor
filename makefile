.PHONY: test

apis:
	oapi-codegen --config codegen/config.yaml api/api.yaml | gofmt > api/generated.go

test: backend-test feature-test

test-backend:
	$(eval export TEMPDIR := $(shell mktemp -d))
	CGO_ENABLED=0 go build -ldflags="-w -s" -o $(TEMPDIR)/monitor
	$(TEMPDIR)/monitor & echo "$$!" > $(TEMPDIR)/PID

cleanup-test-backend:
	kill -9 `cat $(TEMPDIR)/PID`
	rm -rf $(TEMPDIR)

feature-test: test-backend
	NODE_ENV=test npm install
	NODE_ENV=test npm run build
	NODE_ENV=test npm run test
	@$(MAKE) --no-print-directory cleanup-test-backend

backend-test:
	go test ./... -race

docker:
	docker build -t nutsfoundation/nuts-monitor:latest .
