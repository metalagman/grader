fail:
	@echo "Please specify proper PART_ID"
	exit 1

lint:
	@echo "Running linter checks"
	golangci-lint run

test:
	@echo "Running $(PART_ID) tests"
	bash run-tests.sh $(PART_ID)
