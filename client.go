package main

import (
	proto "github.com/DarRo9/proto5"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"log"
)

func main() {

	var conn *grpc.ClientConn
	conn, err := grpc.Dial(":9000", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %s", err)
	}
	defer conn.Close()

	c := proto.NewDelAccClient(conn)

	response, err := c.Do(context.Background(), &proto.Request{Message: "Удоли"})
	if err != nil {
		log.Fatalf("Error when calling SayHello: %s", err)
	}
	log.Printf("Response from server: %s", response.Message)

}
