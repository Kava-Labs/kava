.PHONY: docker-build
docker-build:
	docker build -t kava/kava:local ../..

.PHONY: test-e2e
test-e2e:
	export E2E_KAVA_FUNDED_ACCOUNT_MNEMONIC='season bone lucky dog depth pond royal decide unknown device fruit inch clock trap relief horse morning taxi bird session throw skull avocado private'; \
	go test -failfast -count=1 -v ./...
