.PHONY: docker-build
docker-build:
	docker build -t kava/kava:local ../..

.PHONY: e2e-test
e2e-test:
	go test -v ./...
