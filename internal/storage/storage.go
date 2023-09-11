package storage

import (
	"database/sql"
	"time"
	_ "github.com/lib/pq"
	"orders/internal/config"
)

type StructJsonWb struct {
	OrderUid    string `json:"order_uid"`
	TrackNumber string `json:"track_number"`
	Entry       string `json:"entry"`
	Delivery    struct {
		Name    string `json:"name"`
		Phone   string `json:"phone"`
		Zip     string `json:"zip"`
		City    string `json:"city"`
		Address string `json:"address"`
		Region  string `json:"region"`
		Email   string `json:"email"`
	} `json:"delivery"`
	Payment struct {
		Transaction  string `json:"transaction"`
		RequestId    string `json:"request_id"`
		Currency     string `json:"currency"`
		Provider     string `json:"provider"`
		Amount       int    `json:"amount"`
		PaymentDt    int    `json:"payment_dt"`
		Bank         string `json:"bank"`
		DeliveryCost int    `json:"delivery_cost"`
		GoodsTotal   int    `json:"goods_total"`
		CustomFee    int    `json:"custom_fee"`
	} `json:"payment"`
	Items []struct {
		ChrtId      int    `json:"chrt_id"`
		TrackNumber string `json:"track_number"`
		Price       int    `json:"price"`
		Rid         string `json:"rid"`
		Name        string `json:"name"`
		Sale        int    `json:"sale"`
		Size        string `json:"size"`
		TotalPrice  int    `json:"total_price"`
		NmId        int    `json:"nm_id"`
		Brand       string `json:"brand"`
		Status      int    `json:"status"`
	} `json:"items"`
	Locale            string    `json:"locale"`
	InternalSignature string    `json:"internal_signature"`
	CustomerId        string    `json:"customer_id"`
	DeliveryService   string    `json:"delivery_service"`
	Shardkey          string    `json:"shardkey"`
	SmId              int       `json:"sm_id"`
	DateCreated       time.Time `json:"date_created"`
	OofShard          string    `json:"oof_shard"`
}

//хэш-таблица для кэширования информации
var CashOrders map[string][]byte 

type orders struct {
	Uid  string
	Info []byte
}

//Функция для кеширования информации
func CacheUP(db *sql.DB) error {

	rows, err := db.Query("SELECT * from orders_table")
	if err != nil {
		return err
	}

	var items []orders
	for rows.Next() {
		post := orders{}
		err = rows.Scan(&post.Uid, &post.Info)
		if err != nil {
			return err
		}
		items = append(items, post)
	}
	err = rows.Close()
	if err != nil {
		return err
	}
	for _, i := range items {
		CashOrders[i.Uid] = i.Info

	}
	return nil

}

func ConnectToDb(conf *config.Config) (*sql.DB, error) {

	db, err := sql.Open(conf.DriverName, conf.DSN)
	if err != nil {
		return nil, err
	}
	
	//SetMaxOpenConns устанавливает максимальное количество открытых подключений к базе данных
	db.SetMaxOpenConns(10)

	//Ping проверяет, существует ли соединение с базой данных, и при необходимости устанавливает соединение
	err = db.Ping()
	if err != nil {
		return nil, err
	}
	return db, nil
}

func Insert(uid, jsonOrder string, db *sql.DB) error {
	_, err := db.Exec(
		"INSERT INTO orders_table (uid, json_order) VALUES ($1, $2 )", uid, jsonOrder)
	if err != nil {
		return err
	}
	return nil
}


