release="1.0.4"

.PHONY: up
up:
	docker compose up -d

.PHONY: down
down:
	docker compose down

.PHONY: run
run:
	go run cmd/app/main.go

.PHONY: build
build:
ifdef repo
	docker build --platform linux/amd64 -t ${repo}/messagio:${release} .
else
	@echo "ERROR: Repository is required. Try 'make repo=XXX build'"
endif

.PHONY: push
push:
ifdef repo
	docker push ${repo}/messagio:${release}
else
	@echo "ERROR: Repository is required. Try 'make repo=XXX push'"
endif

.PHONY: swag
swag:
	swag init --parseDependency --parseInternal -g cmd/app/main.go  -o ./internal/app/docs

.DEFAULT_GOAL := up
