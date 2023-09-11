all :
	docker-compose up

serv:
	go run cmd/main.go

example:
	go run publisher/publish.go  sent.json
