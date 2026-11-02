.PHONY: build test dev-config

build:
	mkdir -p bin
	go build -o bin/terraform-provider-cato-refactor .

test:
	go test ./...

dev-config: build
	printf 'provider_installation {\n  dev_overrides {\n    "catonetworks/cato-refactor" = "%s"\n  }\n  direct {}\n}\n' '$(CURDIR)/bin' > .terraformrc
