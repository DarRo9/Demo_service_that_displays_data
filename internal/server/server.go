package server

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"orders/internal/config"
	storage "orders/internal/storage"
	model "orders/internal/model"
)

func handler(w http.ResponseWriter, r *http.Request) {

	html, err := ioutil.ReadFile("Templates/index.html")
	if err != nil {
		log.Fatal(err)
	}
	_, err = fmt.Fprintln(w, string(html))
	if err != nil {
		log.Fatal(err)
	}

	tmpl := template.Must(template.ParseFiles("Templates/inf.html"))
	var str model.Order

	key := r.FormValue("q")

	if key != "" {
		res, ok := (storage.CashOrders)[key]
		fmt.Print(res)
		if ok {
			str2, _ := json.Marshal((storage.CashOrders)[key])
			fmt.Print(str2)
			err = json.Unmarshal(str2, &str)
			if err != nil {
				log.Println(err)
			}
			err = tmpl.Execute(w, str)
			if err != nil {
				log.Println(err)
			}
		}
	}
}

func ServerLaunch(conf *config.Config) error {
	//HandleFunc добавляет обработчики в DefaultServeMux
	http.HandleFunc("/", handler)

	log.Println("starting server at :8080")

	//ListenAndServe запускает HTTP-сервер с заданным адресом и обработчиком. 
	err := http.ListenAndServe(conf.Port, nil)
	if err != nil {
		return err
	}
	return nil
}
