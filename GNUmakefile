BUILD_ENV=CGO_ENABLED=0
MODULE_NAME=github.com/scalr/go-scalr
test:
	$(BUILD_ENV)  go test -v $(TESTARGS) -timeout=180s -covermode atomic -coverprofile=covprofile $(MODULE_NAME)

.PHONY: test
