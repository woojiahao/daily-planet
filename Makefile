GOOS=linux
GOARCH=amd64
GOLANG_MIGRATE := $(shell command -v migrate 2> /dev/null)
MIGRATION_NAME := $(NAME)
MIGRATION_UP := $(UP)
MIGRATION_DOWN := $(DOWN)

.PHONY: install linux-build run migrate-new

install:
	go mod download

build:
	go build -o ./out/daily-planet **/*.go

build-linux:
	env GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o ./out/daily-planet **/*.go

run:
	go run main.go

migrate-new:
ifndef GOLANG_MIGRATE
	$(error "golang-migrate is not available. please install it.")
endif
ifndef MIGRATION_NAME
	$(error "--NAME parameter not provided for migration file.")
endif
	migrate create -ext sql -dir db/migrations $(MIGRATION_NAME)

migrate-up:
	migrate -path db/migrations -database "sqlite3://daily_planet.db" up $(MIGRATION_UP)


migrate-down:
	migrate -path db/migrations -database "sqlite3://daily_planet.db" down $(MIGRATION_DOWN)
