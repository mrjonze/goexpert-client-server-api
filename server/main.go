package main

import (
	"context"
	"database/sql"
	"encoding/json"
	_ "github.com/mattn/go-sqlite3"
	"io"
	"log"
	"net/http"
	"text/template"
	"time"
)

type Price struct {
	Usdbrl struct {
		Code       string `json:"code"`
		Codein     string `json:"codein"`
		Name       string `json:"name"`
		High       string `json:"high"`
		Low        string `json:"low"`
		VarBid     string `json:"varBid"`
		PctChange  string `json:"pctChange"`
		Bid        string `json:"bid"`
		Ask        string `json:"ask"`
		Timestamp  string `json:"timestamp"`
		CreateDate string `json:"create_date"`
	} `json:"USDBRL"`
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Por favor, use a rota /cotacao"))
	})
	http.HandleFunc("/cotacao", SearchPriceHandler)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}

func SearchPriceHandler(w http.ResponseWriter, r *http.Request) {
	price, err := SearchPrice()

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Erro ao buscar preço."))
		log.Println(err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	tmpl := template.Must(template.New("PriceResponseTemplate").Parse(`{"bid": "{{.Usdbrl.Bid}}" }`))

	err = tmpl.Execute(w, price)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Erro ao fazer o parse da resposta."))
		log.Println(err)
		return
	}
}

func SearchPrice() (*Price, error) {
	url := "https://economia.awesomeapi.com.br/json/last/USD-BRL"

	req, err := http.Get(url)

	/*
		Fiz a tentativa de usar o contexto para cancelar a requisição após 200ms, mas não consegui fazer funcionar.

		2024/03/31 22:21:59 http: panic serving [::1]:52115: runtime error: invalid memory address or nil pointer dereference
		goroutine 19 [running]:
		net/http.(*conn).serve.func1()
		        /usr/local/go/src/net/http/server.go:1854 +0xb0
		panic({0x10492a040, 0x104b4cd40})
		        /usr/local/go/src/runtime/panic.go:890 +0x258
		main.SearchPrice()
	*/

	/*
		ctx2, cancel2 := context.WithTimeout(context.Background(), 200*time.Millisecond)
		defer cancel2()
		req, err := http.NewRequestWithContext(ctx2, "GET", url, nil)
	*/

	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer req.Body.Close()

	res, err := io.ReadAll(req.Body)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	var price Price
	err = json.Unmarshal(res, &price)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	dsn := "sqlite3:database.db"

	db, err := sql.Open("sqlite3", dsn)
	defer db.Close()

	if err != nil {
		log.Println(err)
		return nil, err
	}

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS price (bid TEXT, code TEXT, codein TEXT, name TEXT, high TEXT, low TEXT, varBid TEXT, pctChange TEXT, ask TEXT, create_date TEXT)")
	if err != nil {
		log.Println(err)
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	_, err = db.ExecContext(ctx, "INSERT INTO price (bid, code, codein, name, high, low, varBid, pctChange, ask, create_date) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		price.Usdbrl.Bid,
		price.Usdbrl.Code,
		price.Usdbrl.Codein,
		price.Usdbrl.Name,
		price.Usdbrl.High,
		price.Usdbrl.Low,
		price.Usdbrl.VarBid,
		price.Usdbrl.PctChange,
		price.Usdbrl.Ask,
		price.Usdbrl.CreateDate)

	if err != nil {
		log.Println(err)
		return nil, err
	}

	return &price, nil
}
