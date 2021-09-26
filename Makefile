up:
	docker-compose up --build

local-app:
	./run-local-app.sh

test: migrate-up
	go test ./tests -test.v

migrate-up:
	migrate -path datastore/postgres/migrations -database "postgresql://postgres:postgres@localhost:5432/postgres?sslmode=disable" -verbose up

migrate-down:
	migrate -path datastore/postgres/migrations -database "postgresql://postgres:postgres@localhost:5432/postgres?sslmode=disable" -verbose down --all
