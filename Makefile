# Sources (package paths)
GLCORE_PKG=./cmd/glcore
GLCLI_PKG=./cmd/glcli

# Destionations
DIR_DEST=bin/
GLCORE_NAME=$(DIR_DEST)glcore
GLCLI_NAME=$(DIR_DEST)glcli

# Install destination
INSTALL_PATH=/usr/local/bin
INSTALLED_GLCORE=$(INSTALL_PATH)/glcore
INSTALLED_GLCLI=$(INSTALL_PATH)/glcli

# Compiler flags
GO_FLAGS=-o

.PHONY: all build-glcore build-glcli run-glcore run-glcli clean clean-glcore clean-glcli fclean re install uninstall test test-coverage test-verbose test-race

all: build-glcore build-glcli

$(DIR_DEST):
	mkdir -p $(DIR_DEST)

build-glcore:
	go build $(GO_FLAGS) $(GLCORE_NAME) $(GLCORE_PKG)

build-glcli:
	go build $(GO_FLAGS) $(GLCLI_NAME) $(GLCLI_PKG)

run-glcore: build-glcore
	./$(GLCORE_NAME)

run-glcli: build-glcli
	./$(GLCLI_NAME)

clean: clean-glcore clean-glcli
clean-glcore:
	rm -f $(GLCORE_NAME)
clean-glcli:
	rm -f $(GLCLI_NAME)
fclean: clean
	rm -fr $(DIR_DEST)

install: build-glcore build-glcli
	sudo mkdir -p $(INSTALL_PATH)
	sudo install -m 755 $(GLCORE_NAME) $(INSTALLED_GLCORE)
	sudo install -m 755 $(GLCLI_NAME) $(INSTALLED_GLCLI)
	@echo "glcore and glcli installed in $(INSTALL_PATH)"
uninstall:
	sudo rm -f $(INSTALLED_GLCORE) $(INSTALLED_GLCLI)
	@echo "glcore and glcli removed from $(INSTALL_PATH)"
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
