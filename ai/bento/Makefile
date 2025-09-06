
DIRS := $(wildcard */)

dirs:
	echo "$(DIRS)"

%/serve:
	@echo "Running serve for directory: $(patsubst %/serve,%,$@)"
	@cd $(patsubst %/serve,%,$@) && uv run bentoml serve

%/test:
	@echo "Running serve for directory: $(patsubst %/test,%,$@)"
	@cd $(patsubst %/test,%,$@) && ./test.sh
