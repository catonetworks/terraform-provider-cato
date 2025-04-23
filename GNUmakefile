WEBSITE_REPO=github.com/hashicorp/terraform-website
HOSTNAME=registry.terraform.io
NAMESPACE=terraform-providers
PKG_NAME=cato
BINARY=terraform-provider-${PKG_NAME}
# Whenever bumping provider version, please update the version in cato/client.go (line 27) as well.
VERSION=0.0.20

# Mac Intel Chip
# OS_ARCH=darwin_amd64
# For Mac M1 Chip
OS_ARCH=darwin_arm64
# OS_ARCH=linux_amd64

default: install

build:
	export GO111MODULE="on"
	go mod vendor
	go build -o ${BINARY}

install: build
	mkdir -p ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${PKG_NAME}/${VERSION}/${OS_ARCH}
	mv ${BINARY} ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${PKG_NAME}/${VERSION}/${OS_ARCH}

clean: install
	go clean -cache -modcache -i -r