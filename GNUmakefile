default: testacc

setup:
	cd ~ && go get gotest.tools/gotestsum && cd -

# Run acceptance tests
.PHONY: testacc
testacc:
	TF_ACC=1 gotestsum ./... -timeout 120m
