BUILD_ENV=CGO_ENABLED=0
PKGS := $(shell go list ./... | grep -v '/v2')

test:
	$(BUILD_ENV) go test -v $(TESTARGS) -timeout=360s -covermode atomic -coverprofile=covprofile $(PKGS)

.PHONY: test
