package main

import (
	"fmt"
	"github.com/nats-io/stan.go"
	"log"
	"os"
	"orders/internal/config"
)

func main() {
    // config.GetConfig() возвращает структуру с информацией о переменных среды
	conf, err := config.GetConfig() 
	if err != nil {
		log.Fatal("error during config downloading: ", err)
	}
	sc, err := stan.Connect(conf.StanClusterId, "clientID2")
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		err = sc.Close()
		if err != nil {
			fmt.Println(err)
		}
	}()

	mod, err := os.ReadFile(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	err = sc.Publish(conf.Subject, mod)
	if err != nil {
		log.Println(err)
	}
}
