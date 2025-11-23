up:
	docker compose up --build -d

down:
	docker compose down

clean:
	docker compose down -v

int-test: 
	docker run --rm --network=host tests:latest

load-test:
	curl -X POST "http://localhost:8080/team/add" \
	-H "Content-Type: application/json" \
	-d '{"team_name":"test-team","members":[{"user_id":"user1","username":"Test User","is_active":true}]}'
	
	bombardier -c 20 -d 15s -l "http://localhost:8080/team/get?team_name=test-team"

test:
	make clean
	make up
	make int-test
	make load-test
	make clean
	@echo "test finished"

lint:
	golangci-lint run -v ./...

tools:
	go install github.com/codesenberg/bombardier@latest
# 	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate
# 	go install golang.org/x/tools/cmd/goimports@latest
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b $$(go env GOPATH)/bin v2.4.0
