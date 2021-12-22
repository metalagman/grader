fail:
	@echo "Please specify proper PART_ID"
	exit 1

lint:
	@echo "Running linter checks"
	golangci-lint run

test:
	@echo "Running $(PART_ID) tests"
	@cp /app/submission/*.go /app/$(PART_ID) 2> /dev/null || true
	@bash run-tests.sh $(PART_ID)
