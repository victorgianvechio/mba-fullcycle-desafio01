package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

type ExchangeRate struct {
	Valor string `json:"bid"`
}

const MAX_API_TIMEOUT = 300 * time.Millisecond
const API_URL = "http://localhost:8080/cotacao"

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), MAX_API_TIMEOUT)
	defer cancel()

	// Prepara a request
	req, err := http.NewRequestWithContext(ctx, "GET", API_URL, nil)
	if err != nil {
		log.Printf("erro ao montar o request: %s", err.Error())
	}

	// executa a request
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("erro ao executar o request: %s", err.Error())
		return
	}
	defer res.Body.Close()

	f, err := os.Create("cotacao.txt")
	if err != nil {
		log.Printf("erro ao criar arquivo: %s", err.Error())
		return
	}

	// Decodifica o JSON para acessar o campo "bid"
	var result ExchangeRate
	err = json.NewDecoder(res.Body).Decode(&result)
	if err != nil {
		log.Printf("Erro ao decodificar JSON: %s", err.Error())
		return
	}

	fmt.Println("\nValor do bid:", result.Valor)

	_, err = f.WriteString(fmt.Sprintf("Dólar: %s\n", result.Valor))
	if err != nil {
		log.Printf("erro ao escrever no arquivo: %s", err.Error())
		return
	}
}
