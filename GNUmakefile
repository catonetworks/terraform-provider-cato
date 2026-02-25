WEBSITE_REPO=github.com/hashicorp/terraform-website
HOSTNAME=registry.terraform.io
NAMESPACE=catonetworks
PKG_NAME=cato
BINARY=terraform-provider-${PKG_NAME}
# Whenever bumping provider version, please update the version in cato/client.go (line 27) as well.
VERSION=0.0.64

# Mac Intel Chip
# OS_ARCH=darwin_amd64
# For Mac M1 Chip
OS_ARCH=darwin_arm64
# OS_ARCH=linux_amd64

# Directory paths
PLUGINS_DIR=${HOME}/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${PKG_NAME}/${VERSION}/${OS_ARCH}
MIRROR_DIR=${HOME}/.terraform.d/mirror/${HOSTNAME}/${NAMESPACE}/${PKG_NAME}/${VERSION}/${OS_ARCH}

default: install

build:
	export GO111MODULE="on"
	go mod vendor
	go build -o ${BINARY}

install: build
	mkdir -p ${PLUGINS_DIR}
	cp ${BINARY} ${PLUGINS_DIR}/${BINARY}
	chmod 755 ${PLUGINS_DIR}/${BINARY}
	rm -f ${BINARY}
	@echo 'provider_installation {' > ~/.terraformrc-dev
	@echo '  dev_overrides {' >> ~/.terraformrc-dev
	@echo '    "${NAMESPACE}/${PKG_NAME}" = "${PLUGINS_DIR}"' >> ~/.terraformrc-dev
	@echo '  }' >> ~/.terraformrc-dev
	@echo '  direct {}' >> ~/.terraformrc-dev
	@echo '}' >> ~/.terraformrc-dev
	@echo "✓ Provider v${VERSION} installed to plugins directory"
	@echo "  Run 'tfsync dev' to use dev_overrides mode"

install-mirror: build
	@# Install to plugins directory (for dev mode with dev_overrides)
	mkdir -p ${PLUGINS_DIR}
	cp ${BINARY} ${PLUGINS_DIR}/${BINARY}
	chmod 755 ${PLUGINS_DIR}/${BINARY}
	@# Create proper filesystem mirror structure for terraform test
	@# This uses the exact structure that terraform init --platform=... would create
	mkdir -p ${MIRROR_DIR}
	@# Copy with the version-tagged name that terraform expects
	cp ${BINARY} ${MIRROR_DIR}/${BINARY}_v${VERSION}
	chmod 755 ${MIRROR_DIR}/${BINARY}_v${VERSION}
	@# Also symlink the unversioned name for compatibility
	ln -sf ${BINARY}_v${VERSION} ${MIRROR_DIR}/${BINARY}
	@# Clean up local binary
	rm -f ${BINARY}
	@# Update terraformrc configs
	@echo 'provider_installation {' > ~/.terraformrc-dev
	@echo '  dev_overrides {' >> ~/.terraformrc-dev
	@echo '    "${NAMESPACE}/${PKG_NAME}" = "${PLUGINS_DIR}"' >> ~/.terraformrc-dev
	@echo '  }' >> ~/.terraformrc-dev
	@echo '  direct {}' >> ~/.terraformrc-dev
	@echo '}' >> ~/.terraformrc-dev
	@# For mirror mode, use filesystem_mirror
	@echo 'provider_installation {' > ~/.terraformrc-mirror
	@echo '  filesystem_mirror {' >> ~/.terraformrc-mirror
	@echo '    path    = "${HOME}/.terraform.d/mirror"' >> ~/.terraformrc-mirror
	@echo '    include = ["${NAMESPACE}/${PKG_NAME}"]' >> ~/.terraformrc-mirror
	@echo '  }' >> ~/.terraformrc-mirror
	@echo '  direct {' >> ~/.terraformrc-mirror
	@echo '    exclude = ["${NAMESPACE}/${PKG_NAME}"]' >> ~/.terraformrc-mirror
	@echo '  }' >> ~/.terraformrc-mirror
	@echo '}' >> ~/.terraformrc-mirror
	@echo "✓ Provider v${VERSION} installed to both plugins and mirror directories"
	@echo "  Run 'tfsync dev' for fast development (incompatible with terraform test)"
	@echo "  Run 'tfsync mirror' for testing (compatible with terraform test)"

sync-provider:
	@# Create .terraform/providers directory structure to satisfy catocli checks
	@# Run this in any Terraform project directory where you want to use catocli
	@TARGET_DIR=$${PROJECT_DIR:-.}; \
	if [ ! -d "$$TARGET_DIR/.terraform" ]; then \
		echo "Error: Must run from a Terraform project directory"; \
		exit 1; \
	fi; \
	mkdir -p $$TARGET_DIR/.terraform/providers/${HOSTNAME}/${NAMESPACE}/${PKG_NAME}/${VERSION}/${OS_ARCH}; \
	ln -sf ${PLUGINS_DIR}/${BINARY} \
		$$TARGET_DIR/.terraform/providers/${HOSTNAME}/${NAMESPACE}/${PKG_NAME}/${VERSION}/${OS_ARCH}/${BINARY}; \
	echo "✓ Provider symlinked to $$TARGET_DIR/.terraform/providers"

clean: install
	go clean -cache -modcache -i -r
