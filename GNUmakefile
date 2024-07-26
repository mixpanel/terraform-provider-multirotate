default: testacc

# Run acceptance tests
.PHONY: testacc
testacc:
	TF_ACC=1 go test ./... -v $(TESTARGS) -timeout 120m

# Flake test acceptance tests
.PHONY: testaccflake
testaccflake:
	TF_ACC=1 go test ./... -v $(TESTARGS) -timeout 120m -count=10
