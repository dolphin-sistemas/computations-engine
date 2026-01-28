.PHONY: help wasm server server-win all clean

help:
	@echo "Targets:"
	@echo "  wasm        Build WASM (client/wasm/order_engine.wasm + wasm_exec.js)"
	@echo "  server      Build server for current OS (pdv-jsonlogic.exe on Windows, pdv-jsonlogic on *nix)"
	@echo "  server-win  Cross-compile Windows amd64 server exe (pdv-jsonlogic.exe)"
	@echo "  all         wasm + server"
	@echo "  clean       Remove build outputs"

wasm:
	@bash ./scripts/build-wasm.sh

# Build for the current OS/arch
server:
	@go build -o pdv-jsonlogic$(if $(filter Windows_NT,$(OS)),.exe,) ./cmd/api

# Cross-compile Windows amd64 (useful from WSL/Linux)
server-win:
	@GOOS=windows GOARCH=amd64 go build -o pdv-jsonlogic.exe ./cmd/api
	
clean:
	@rm -f pdv-jsonlogic.exe pdv-jsonlogic pdv-jsonlogic-linux
	@rm -rf dist

run:
	go run cmd/api/main.go