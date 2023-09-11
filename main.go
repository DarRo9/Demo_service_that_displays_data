package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"orders/internal/Nats"
	server "orders/internal/server"
	"orders/internal/config"
	storage "orders/internal/storage"
)



func main() {

	
	// config.GetConfig() возвращает структуру с информацией о переменных среды
	conf, err := config.GetConfig()
	if err != nil {
		log.Fatal("error during config downloading: ", err)
	}
	
	db, err := storage.ConnectToDb(conf)
	if err != nil {
		log.Fatal(err)
	}

	//Связь и подписка на Nats
	sc, sub, err := Nats.GetSub(conf, db, &storage.CashOrders)
	if err != nil {
		log.Println(err)
	}

	go func() {
		//server.ServerLaunch(conf) запускает HTTP-сервер с заданным адресом и обработчиком, добавляет обработчики в DefaultServeMux
		err = server.ServerLaunch(conf)
		if err != nil {
			log.Fatal(err)
		}
	}()

	quit := make(chan os.Signal, 1)

	/*Notify заставляет сигнал пакета ретранслировать входящие сигналы в канал.
	Иногда нам хотелось бы, чтобы наши программы на Go интеллектуально обрабатывали сигналы Unix. 
	Например, мы можем захотеть, чтобы сервер корректно завершил работу при получении SIGTERM, или инструмент командной строки остановил обработку ввода, если он получил SIGINT. 
	Вот как обрабатывать сигналы в Go с каналами.
	Уведомление о выходе сигнала работает путем отправки значений os.Signal в канал. 
	Мы создадим канал для получения этих уведомлений (мы также создадим канал, чтобы уведомить нас, когда программа может выйти).
	Эта горутина выполняет блокировку приема сигналов. 
	Когда она получит его, то распечатает его, а затем уведомит программу, что она может быть завершена.
	Программа будет ждать здесь, пока не получит ожидаемый сигнал (как указано в приведенной выше процедуре, отправляющей значение в done), и затем завершится.
	Когда мы запустим эту программу, она заблокирует ожидание сигнала. 
	Набрав ctrl-C (который терминал показывает как ^C), мы можем послать сигнал SIGINT, в результате чего программа напечатает interrupt и затем выйдет.*/
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit
	log.Println("app Shutting down")
	if err := db.Close(); err != nil {
		log.Println(err)
	}
	if sc != nil {
		if err = sc.Close(); err != nil {
			log.Println(err)
		}
	}
	if sub != nil {
		if err = sub.Unsubscribe(); err != nil {
			log.Println(err)
		}
	}
}
