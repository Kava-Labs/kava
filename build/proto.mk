.PHONY: proto-lint check-proto-lint
proto-lint check-proto-lint: install-build-deps
	@echo "Linting proto file"
	@$(BUF) lint

.PHONY: proto-gen
proto-gen: install-build-deps
	@echo "Generating go proto files"
	@$(BUF) generate --template proto/buf.gen.gogo.yaml proto
	@cp -r out/github.com/kava-labs/kava/* ./
	@rm -rf out/github.com

.PHONY: check-proto-gen
check-proto-gen: proto-gen ## Return error code 1 if proto gen changes files
	@git diff --exit-code **/*.pb.go > /dev/null || (echo "Protobuf generated go files are not up to date! Please run \`make proto-gen\`."; exit 1)

.PHONY: proto-gen-doc
proto-gen-doc: install-build-deps
	@echo "Generating proto doc"
	@$(BUF) generate --template proto/buf.gen.doc.yaml proto

.PHONY: check-proto-gen-doc
check-proto-gen-doc: proto-gen-doc ## Return error code 1 if proto gen changes files
	@git diff --exit-code docs/core/proto-docs.md > /dev/null || (echo "Protobuf doc is not up to date! Please run \`make proto-gen-doc\`."; exit 1)

.PHONY: proto-gen-swagger
proto-gen-swagger: install-build-deps
	@echo "Generating proto swagger"
	@$(BUF) generate --template proto/buf.gen.swagger.yaml proto
	@$(SWAGGER_COMBINE) client/docs/config.json -o client/docs/swagger-ui/swagger.yaml -f yaml --continueOnConflictingPaths true --includeDefinitions true
	@rm -rf out/swagger

.PHONY: check-proto-gen-swagger
check-proto-gen-swagger: proto-gen-swagger ## Return error code 1 if proto gen changes files
	@git diff --exit-code client/docs/swagger-ui/swagger.yaml > /dev/null || (echo "Protobuf swagger is not up to date! Please run \`make proto-gen-swagger\`."; exit 1)

.PHONY: proto-format
	@echo "Formatting proto files"
proto-format: install-build-deps
	@$(BUF) format -w proto

.PHONY: check-proto-format
check-proto-format: proto-format
	@git diff --exit-code proto/**/*.proto > /dev/null || (echo "Protobuf format is not up to date! Please run \`make proto-format\`."; exit 1)

BUF_CHECK_BREAKING_AGAINST ?= ref=HEAD~1
BUF_CHECK_BREAKING_AGAINST_REMOTE ?= branch=$(GIT_BRANCH),$(BUF_CHECK_BREAKING_AGAINST)

.PHONY: check-proto-breaking
check-proto-breaking: install-build-deps
	@echo "Checking for proto backward compatibility"
	@$(BUF) breaking --against '.git#$(BUF_CHECK_BREAKING_AGAINST)'

.PHONY: check-proto-breaking-remote
check-proto-breaking-remote: install-build-deps
	@echo "Checking for proto backward compatibility"
	$(BUF) breaking --against '$(HTTPS_GIT)#$(BUF_CHECK_BREAKING_AGAINST_REMOTE)'

.PHONY: proto-gen-all
proto-gen-all: proto-gen proto-gen-doc proto-gen-swagger

.PHONY: proto-all
proto-all: proto-update-deps proto-lint proto-format check-proto-breaking proto-gen-all
