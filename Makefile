up: ## Database migration up
	@go run cmd/migrate/main.go up

down: ## Database migration down
	@go run cmd/migrate/main.go down