package storage

import (
	"database/sql"
	_ "github.com/lib/pq"
	"orders/internal/config"
	model "orders/internal/model"
	"context"
	"fmt"
	
)

type orders struct {
	Uid  string
	Info []byte
}

const createOrder = `INSERT INTO orders (order_uid, track_number, entry, locale, internal_signature, customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`
const createDelivery = `INSERT INTO	deliveries (order_uid, name, phone, zip, city, address, region, email) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
const createPayment = `INSERT INTO payments	(order_uid, transaction, request_id, currency, provider, amount, payment_dt, bank, delivery_cost, goods_total, custom_fee) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`
const createItem = `INSERT INTO	items (order_uid, chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`

type Storage struct {
	db *sql.DB
}

func NewStorage(db *sql.DB) *Storage {
	return &Storage{
		db: db,
	}
}

//хэш-таблица для кэширования информации
var CashOrders map[string]model.Order 

func (p *Storage) CreateOrder(ctx context.Context, order *model.Order) (string, error) {
	tx, err := p.db.Begin()
	if err != nil {
		return "", fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	_, err = tx.Exec(createOrder,
		order.OrderUID, order.TrackNumber, order.Entry, order.Locale,
		order.InternalSignature, order.CustomerID, order.DeliveryService, order.Shardkey,
		order.SmID, order.DateCreated, order.OofShard)
	if err != nil {
		return "", fmt.Errorf("error when adding an entry to orders: %w", err)
	}

	_, err = tx.Exec(createDelivery,
		order.OrderUID, order.Delivery.Name, order.Delivery.Phone, order.Delivery.Zip,
		order.Delivery.City, order.Delivery.Adress, order.Delivery.Region, order.Delivery.Email)
	if err != nil {
		return "", fmt.Errorf("error when adding an entry to deliveries: %w", err)
	}

	_, err = tx.Exec(createPayment,
		order.OrderUID, order.Payment.Transaction, order.Payment.RequestID, order.Payment.Currency,
		order.Payment.Provider, order.Payment.Amount, order.Payment.PaymentDT, order.Payment.Bank,
		order.Payment.DeliveryCost, order.Payment.GoodsTotal, order.Payment.CustomFee)
	if err != nil {
		return "", fmt.Errorf("error when adding an entry to payments: %w", err)
	}

	for i := range order.Items {
		_, err = tx.Exec(createItem,
			order.OrderUID, order.Items[i].ChrtID, order.Items[i].TrackNumber, order.Items[i].Price,
			order.Items[i].Rid, order.Items[i].Name, order.Items[i].Sale, order.Items[i].Size,
			order.Items[i].TotalPrice, order.Items[i].NmID, order.Items[i].Brand, order.Items[i].Status)
		if err != nil {
			return "", fmt.Errorf("error when adding an entry to items: %w", err)
		}
	}

	err = tx.Commit()
	if err != nil {
		return "", fmt.Errorf("transaction confirmation error: %w", err)
	}

	return order.OrderUID, nil
}

func (p *Storage) RecoverCache(ctx context.Context, CashOrders *map[string]model.Order) error {
	tx, err := p.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	// Загрузка заказов в кэш
	rows, err := tx.Query(`SELECT * FROM orders`)
	if err != nil {
		return fmt.Errorf("request execution error (getOrdersQuery): %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var order model.Order
		err := rows.Scan(&order.OrderUID, &order.TrackNumber, &order.Entry, &order.Locale, &order.InternalSignature,
			&order.CustomerID, &order.DeliveryService, &order.Shardkey, &order.SmID, &order.DateCreated,
			&order.OofShard)
		if err != nil {
			return fmt.Errorf("error when scanning orders lines: %w", err)
		}
		if _, ok := (*CashOrders)[order.OrderUID]; !ok {
			(*CashOrders)[order.OrderUID] = order
		}
	}

	// Загрузка доставок в кэш
	rows, err = tx.Query(`SELECT * FROM deliveries`)
	if err != nil {
		return fmt.Errorf("request execution error (getDeliveiesQuery): %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var delivery model.Delivery
		var orderUID string
		err := rows.Scan(&orderUID, &delivery.Name, &delivery.Phone, &delivery.Zip, &delivery.City, &delivery.Adress,
			&delivery.Region, &delivery.Email)
		if err != nil {
			return fmt.Errorf("error when scanning orders lines: %w", err)
		}
		if val, ok := (*CashOrders)[orderUID]; ok {
			val.Delivery = delivery
			(*CashOrders)[orderUID] = val
		}
	}

	// Загрузка платежей в кэш
	rows, err = tx.Query(`SELECT * FROM payments`)
	if err != nil {
		return fmt.Errorf("request execution error (getPaymentsQuery): %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var payment model.Payment
		var orderUID string
		err := rows.Scan(&orderUID, &payment.Transaction, &payment.RequestID, &payment.Currency, &payment.Provider,
			&payment.Amount, &payment.PaymentDT, &payment.Bank, &payment.DeliveryCost, &payment.GoodsTotal,
			&payment.CustomFee)
		if err != nil {
			return fmt.Errorf("error when scanning orders lines: %w", err)
		}
		if val, ok := (*CashOrders)[orderUID]; ok {
			val.Payment = payment
			(*CashOrders)[orderUID] = val
		}
	}

	// Загрузка элементов в кэш
	rows, err = tx.Query(`SELECT * FROM items`)
	if err != nil {
		return fmt.Errorf("request execution error (getItemsQuery): %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var item model.Item
		var orderUID string
		err := rows.Scan(&orderUID, &item.ChrtID, &item.TrackNumber, &item.Price, &item.Rid, &item.Name, &item.Sale,
			&item.Size, &item.TotalPrice, &item.NmID, &item.Brand, &item.Status)
		if err != nil {
			return fmt.Errorf("error when scanning orders lines: %w", err)
		}
		if val, ok := (*CashOrders)[orderUID]; ok {
			val.Items = append(val.Items, item)
			(*CashOrders)[orderUID] = val
		}
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("transaction confirmation error: %w", err)
	}

	fmt.Println("Data has been successfully uploaded from the database to the cache")

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







