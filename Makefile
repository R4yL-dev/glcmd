# Sources
SRC_DIR=cmd/
GLCORE_SRC=$(SRC_DIR)glcore/main.go

# Destionations
DIR_DEST=bin/
GLCORE_NAME=$(DIR_DEST)glcore

# Install destination
INSTALL_PATH=/usr/local/bin
INSTALLED_NAME=$(INSTALL_PATH)/glcore

# Compiler flags
GO_FLAGS=-o

.PHONY: all build-glcore run-glcore clean clean-glcore fclean re install uninstall test test-coverage test-verbose test-race

all: build-glcore

$(DIR_DEST):
	mkdir -p $(DIR_DEST)

build-glcore:
	go build $(GO_FLAGS) $(GLCORE_NAME) $(GLCORE_SRC)

run-glcore: build-glcore
	./$(GLCORE_NAME)

clean: clean-glcore
clean-glcore:
	rm -f $(GLCORE_NAME)
fclean: clean
	rm -fr $(DIR_DEST)

install: build-glcore
	sudo mkdir -p $(INSTALL_PATH)
	sudo install -m 755 $(GLCORE_NAME) $(INSTALLED_NAME)
	@echo "glcore installed in $(INSTALL_PATH)"
uninstall:
	sudo rm -f $(INSTALLED_NAME)
	@echo "glcore removed from $(INSTALL_PATH)"
reinstall: uninstall install

# Test targets
test:
	go test ./internal/...

test-coverage:
	go test -cover ./internal/...
	go test -coverprofile=coverage.out ./internal/...
	go tool cover -html=coverage.out -o coverage.html

test-verbose:
	go test -v ./internal/...

test-race:
	go test -race ./internal/...

re: fclean all
