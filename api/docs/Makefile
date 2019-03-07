.PHONY: build

V1_YAML = api.yaml


build:
	go run generate.go > generated.yaml
	@echo Building mattermost openapi yaml for v4
	touch $@
	@cat header.yaml > $(V1_YAML)
	@cat generated.yaml >> $(V1_YAML)
	@echo Complete

