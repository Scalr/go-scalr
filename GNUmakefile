BUILD_ENV=CGO_ENABLED=0
PKGS := $(shell go list ./... | grep -v '/v2')
GOTESTSUM=go run gotest.tools/gotestsum@v1.12.0
RERUN_REPORT=rerun-report.txt

test:
	$(BUILD_ENV) $(GOTESTSUM) --format standard-verbose \
		--rerun-fails=2 --rerun-fails-run-root-test \
		--rerun-fails-report=$(RERUN_REPORT) --packages="$(PKGS)" \
		-- $(TESTARGS) -timeout=360s

.PHONY: test
