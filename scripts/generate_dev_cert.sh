#!/bin/bash
set -e

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}=== Development Certificate Generator ===${NC}"
echo ""

# Check if mkcert is installed
if ! command -v mkcert &> /dev/null; then
    echo -e "${YELLOW}mkcert not found. Installing...${NC}"
    
    if [[ "$OSTYPE" == "darwin"* ]]; then
        # macOS
        if command -v brew &> /dev/null; then
            brew install mkcert
        else
            echo "Error: Homebrew is required. Install from https://brew.sh"
            exit 1
        fi
    elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
        # Linux
        echo "Please install mkcert manually:"
        echo "  Ubuntu/Debian: sudo apt install mkcert"
        echo "  Arch: sudo pacman -S mkcert"
        echo "  Or download from: https://github.com/FiloSottile/mkcert"
        exit 1
    else
        echo "Unsupported OS. Please install mkcert from: https://github.com/FiloSottile/mkcert"
        exit 1
    fi
fi

# Install mkcert CA
echo -e "${GREEN}Installing mkcert CA in system trust store...${NC}"
mkcert -install

# Create certs directory
CERTS_DIR="$(dirname "$0")/../services/traefik/certs"
mkdir -p "$CERTS_DIR"

# Generate certificate
echo -e "${GREEN}Generating certificate for *.starterpack.dev and starterpack.dev...${NC}"
cd "$CERTS_DIR"
mkcert -cert-file starterpack.dev.pem -key-file starterpack.dev.key "*.starterpack.dev" "starterpack.dev"

echo ""
echo -e "${GREEN}✓ Certificate generated successfully!${NC}"
echo ""
echo "Files created:"
echo "  - $CERTS_DIR/starterpack.dev.pem"
echo "  - $CERTS_DIR/starterpack.dev.key"
echo ""
echo "Next steps:"
echo "  1. Add to /etc/hosts:"
echo "     sudo sh -c 'echo \"127.0.0.1 api.starterpack.dev\" >> /etc/hosts'"
echo ""
echo "  2. Run docker compose:"
echo "     docker compose up"
echo ""
echo "  3. Access your API at:"
echo "     https://api.starterpack.dev"
echo ""
