#!/bin/bash

set -e

echo "Building WASM (Order Engine)..."

# Criar diretório se não existir
mkdir -p client/wasm

# Definir variáveis para compilação WASM
export GOOS=js
export GOARCH=wasm

# Compilar com otimizações
go build -ldflags="-s -w" -trimpath -o client/wasm/order_engine.wasm ./cmd/wasm

if [ $? -ne 0 ]; then
    echo "❌ Build failed!"
    exit 1
fi

echo ""
echo "Checking WASM size..."
if [ -f "client/wasm/order_engine.wasm" ]; then
    SIZE=$(du -h client/wasm/order_engine.wasm | cut -f1)
    SIZE_BYTES=$(stat -f%z client/wasm/order_engine.wasm 2>/dev/null || stat -c%s client/wasm/order_engine.wasm 2>/dev/null || echo "N/A")
    echo "📦 Size: $SIZE ($SIZE_BYTES bytes)"
fi

echo ""
echo "Copying wasm_exec.js..."

# Tentar encontrar GOROOT
GOROOT=$(go env GOROOT)

# Remover aspas se existirem
GOROOT="${GOROOT%\"}"
GOROOT="${GOROOT#\"}"

# Verificar se wasm_exec.js existe no GOROOT
if [ -f "$GOROOT/misc/wasm/wasm_exec.js" ]; then
    cp "$GOROOT/misc/wasm/wasm_exec.js" "client/wasm/wasm_exec.js"
    echo "✅ wasm_exec.js copied successfully from $GOROOT/misc/wasm/"
else
    echo "⚠️  WARNING: wasm_exec.js not found at $GOROOT/misc/wasm/wasm_exec.js"
    echo ""
    echo "Attempting to download from GitHub..."
    
    # Tentar baixar usando curl ou wget
    if command -v curl &> /dev/null; then
        curl -L -o "client/wasm/wasm_exec.js" "https://raw.githubusercontent.com/golang/go/master/lib/wasm/wasm_exec.js"
    elif command -v wget &> /dev/null; then
        wget -O "client/wasm/wasm_exec.js" "https://raw.githubusercontent.com/golang/go/master/lib/wasm/wasm_exec.js"
    else
        echo "❌ ERROR: Neither curl nor wget found. Please manually download from:"
        echo "   https://raw.githubusercontent.com/golang/go/master/lib/wasm/wasm_exec.js"
        echo "   And save it to: client/wasm/wasm_exec.js"
        exit 1
    fi
    
    if [ -f "client/wasm/wasm_exec.js" ]; then
        echo "✅ wasm_exec.js downloaded successfully from GitHub."
    else
        echo "❌ ERROR: Failed to download wasm_exec.js"
        echo "   Please manually download from:"
        echo "   https://raw.githubusercontent.com/golang/go/master/lib/wasm/wasm_exec.js"
        echo "   And save it to: client/wasm/wasm_exec.js"
        exit 1
    fi
fi

# Limpar variáveis de ambiente
unset GOOS
unset GOARCH

echo ""
echo "✅ WASM build complete!"
echo "📁 Location: client/wasm/order_engine.wasm"
echo ""
