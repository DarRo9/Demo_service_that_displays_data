package Nats

import (
	"database/sql"
	"encoding/json"
	"github.com/nats-io/stan.go"
	"log"
	"orders/internal/config"
	"orders/internal/storage"
	model "orders/internal/model"
	"context"
	"fmt"
)

func initConn(conf *config.Config) (stan.Conn, error) {
	sc, err := stan.Connect(conf.StanClusterId, conf.ClientId, stan.NatsURL(conf.NatsUrl))

	if err != nil {
		return nil, err
	}
	return sc, nil
}

func initSub(conf *config.Config, sc stan.Conn, db *sql.DB, CashOrders *map[string]model.Order) (stan.Subscription, error) {
	//Получение информации в формате json, парсинг в структуру, добавление в таблицы
	var str = model.Order{}

	sub, err := sc.Subscribe(conf.Subject, func(msg *stan.Msg) {
		log.Printf("Received a message: %s\n", string(msg.Data))
		if err := json.Unmarshal(msg.Data, &str); err != nil {
			log.Println(err)
			return
		}
		str2, _ := json.Marshal((*CashOrders)[str.OrderUID])
		str2 = msg.Data
		fmt.Print(str2)
		err := storage.Insert(str.OrderUID, string(msg.Data), db)
		if err != nil {
			log.Println(err)
			return
		}
		ctx := context.Background()
		var order = &model.Order{}
		db2 := storage.NewStorage(db)
		err = db2.RecoverCache(ctx, CashOrders)
		if err != nil {
			log.Println(err)
			return
		}
		_, err = db2.CreateOrder(ctx, order)
		if err != nil {
			log.Println(err)
			return
		}
	}, stan.DurableName(conf.DurableName))

	if err != nil {
		return nil, err
	}
	
	return sub, nil
}

func GetSub(conf *config.Config, db *sql.DB, CashOrders *map[string]model.Order) (stan.Conn, stan.Subscription, error) {
	sc, err := initConn(conf)
	if err != nil {
		return nil, nil, err
	}
	sub, err := initSub(conf, sc, db, CashOrders)
	if err != nil {
		if err = sc.Close(); err != nil {
			log.Println(err)
		}
		return nil, nil, err
	}
	return sc, sub, nil
}
