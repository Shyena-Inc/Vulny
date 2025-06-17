#!/usr/bin/env bash

# Metadata
NAME="Vulny"
VERSION="1.2"
DESCRIPTION="The Multi-Tool Web Vulnerability Scanner."
LONG_DESCRIPTION="A Go-based web vulnerability scanner that runs multiple security tools to identify vulnerabilities in web applications."
URL="https://github.com/Shyena-Inc/Vulny"
AUTHOR="aryanstha4859"

# Installation paths
INSTALL_DIR="/usr/local/bin"
SRC_DIR="$HOME/.vulny"
REPO_URL="https://github.com/Shyena-Inc/Vulny.git"
BINARY_NAME="vulny"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to display metadata
display_metadata() {
    echo -e "${BLUE}Package Metadata:${NC}"
    echo -e "  Name: $NAME"
    echo -e "  Version: $VERSION"
    echo -e "  Description: $DESCRIPTION"
    echo -e "  Long Description: $LONG_DESCRIPTION"
    echo -e "  URL: $URL"
    echo -e "  Author: $AUTHOR"
}

# Function to check if a command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to check dependencies
check_dependencies() {
    echo -e "${BLUE}Checking dependencies...${NC}"

    if ! command_exists go; then
        echo -e "${RED}Error: Go is not installed. Please install Go (version 1.21 or higher).${NC}"
        echo "On Debian/Ubuntu, run: sudo apt install golang"
        exit 1
    fi

    GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
    if [[ "$(printf '%s\n' "$GO_VERSION" "1.21" | sort -V | head -n1)" != "1.21" ]]; then
        echo -e "${RED}Error: Go version $GO_VERSION is too old. Requires 1.21 or higher.${NC}"
        exit 1
    fi

    if ! command_exists git; then
        echo -e "${RED}Error: Git is not installed. Please install Git.${NC}"
        echo "On Debian/Ubuntu, run: sudo apt install git"
        exit 1
    fi

    # Install external tools if missing
    for tool in host nmap wpscan joomscan droopescan sslscan amass nikto; do
        if ! command_exists "$tool"; then
            echo -e "${BLUE}Installing $tool...${NC}"
            case "$tool" in
                host)
                    sudo apt install -y dnsutils ;;
                nmap)
                    sudo apt install -y nmap ;;
                wpscan)
                    sudo apt install -y ruby-dev && sudo gem install wpscan ;;
                joomscan)
                    sudo apt install -y joomscan ;;
                droopescan)
                    sudo apt install -y python3-pip && pip3 install droopescan ;;
                sslscan)
                    sudo apt install -y sslscan ;;
                amass)
                    sudo apt install -y amass ;;
                nikto)
                    sudo apt install -y nikto ;;
                *)
                    echo -e "${RED}Unknown tool: $tool${NC}"
                    exit 1
                    ;;
            esac
        fi
    done

    echo -e "${GREEN}All dependencies satisfied.${NC}"
}

# Function to clone or update the repository
setup_source() {
    echo -e "${BLUE}Setting up source code...${NC}"
    if [[ -d "$SRC_DIR" ]]; then
        echo -e "Repository already exists at $SRC_DIR. Updating..."
        cd "$SRC_DIR" || exit 1
        git pull origin main || {
            echo -e "${RED}Error: Failed to update repository.${NC}"
            exit 1
        }
    else
        echo -e "Cloning repository from $REPO_URL to $SRC_DIR..."
        git clone "$REPO_URL" "$SRC_DIR" || {
            echo -e "${RED}Error: Failed to clone repository.${NC}"
            exit 1
        }
        cd "$SRC_DIR" || exit 1
    fi
}

# Function to build the binary
build_binary() {
    echo -e "${BLUE}Building $BINARY_NAME...${NC}"
    go mod tidy || {
        echo -e "${RED}Error: Failed to tidy Go modules.${NC}"
        exit 1
    }
    go build -o "$BINARY_NAME" || {
        echo -e "${RED}Error: Failed to build $BINARY_NAME.${NC}"
        exit 1
    }
    echo -e "${GREEN}Build completed successfully.${NC}"
}

# Function to install the binary
install_binary() {
    echo -e "${BLUE}Installing $BINARY_NAME to $INSTALL_DIR...${NC}"
    if [[ -f "$BINARY_NAME" ]]; then
        if [[ -w "$INSTALL_DIR" ]]; then
            mv "$BINARY_NAME" "$INSTALL_DIR/$BINARY_NAME"
        else
            echo -e "${BLUE}Sudo permission required to install to $INSTALL_DIR.${NC}"
            sudo mv "$BINARY_NAME" "$INSTALL_DIR/$BINARY_NAME"
        fi
        chmod +x "$INSTALL_DIR/$BINARY_NAME" || {
            echo -e "${RED}Error: Failed to set executable permissions.${NC}"
            exit 1
        }
        echo -e "${GREEN}Installation completed. Run '${BINARY_NAME} -help' to get started.${NC}"
    else
        echo -e "${RED}Error: Binary $BINARY_NAME not found. Build failed.${NC}"
        exit 1
    fi
}

# Main execution
echo -e "${BLUE}Installing $NAME $VERSION...${NC}"
display_metadata
check_dependencies
setup_source
build_binary
install_binary
echo -e "${GREEN}Setup completed successfully!${NC}"