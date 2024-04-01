package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

type Price struct {
	Bid string `json:"bid"`
}

func main() {
	c := http.Client{}
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*300)

	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/cotacao", nil)
	if err != nil {
		log.Fatal(err)
	}

	resp, err := c.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	var price Price
	err = json.Unmarshal(body, &price)

	// Escrita do arquivo com a cotação atual. Pelo enunciado da questão, o arquivo deverá conter somente a cotação atual, não as anteriores.
	file, err := os.Create("cotacao.txt")
	if err != nil {
		log.Fatal(err)
	}

	_, err = file.Write([]byte("Dólar: " + string(price.Bid)))
	if err != nil {
		log.Fatal(err)
	}

	defer file.Close()
}
