SHELL := /bin/sh

.PHONY: verify test frontend-build clean-runtime-artifacts

verify: test frontend-build

test:
	go test ./...

frontend-build:
	pnpm --dir frontend build

clean-runtime-artifacts:
	@if [ -e backend/app ]; then git checkout -- backend/app; fi
	@rm -f backend/switchyard
