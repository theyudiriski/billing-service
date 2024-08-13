run:
	go run ./cmd/

mock:
	mockgen --source=internal/service/loan.go --destination=internal/service/mock/loan.go