# Sources
SRC_DIR=cmd/
GLCMD_SRC=$(SRC_DIR)glcmd/main.go

# Destionations
DIR_DEST=bin/
GLCMD_NAME=$(DIR_DEST)glcmd

# Install destination
INSTALL_PATH=/usr/local/bin
INSTALLED_NAME=$(INSTALL_PATH)/glcmd

# Compiler flags
GO_FLAGS=-o

.PHONY: all build-glcmd run-glcmd clean clean-glcmd fclean re install uninstall

all: build-glcmd

$(DIR_DEST):
	mkdir -p $(DIR_DEST)

build-glcmd:
	go build $(GO_FLAGS) $(GLCMD_NAME) $(GLCMD_SRC)

run-glcmd: build-glcmd
	./$(GLCMD_NAME)

clean: clean-glcmd
clean-glcmd:
	rm -f $(GLCMD_NAME)
fclean: clean
	rm -fr $(DIR_DEST)

install: build-glcmd
	sudo mkdir -p $(INSTALL_PATH)
	sudo install -m 755 $(GLCMD_NAME) $(INSTALLED_NAME)
	@echo "✅ glcmd installed in $(INSTALL_PATH)"
uninstall:
	sudo rm -f $(INSTALLED_NAME)
	@echo "❌ glcmd removed from $(INSTALL_PATH)"
reinstall: uninstall install


re: fclean all
