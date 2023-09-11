all :
	docker-compose up

serv:
	go run cmd/main.go

example:
	go run pub_srcipt/publish.go  sent.json
