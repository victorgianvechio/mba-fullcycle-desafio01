package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Cotacao struct {
	ID    uint `gorm:"primaryKey"`
	Valor string
	Data  time.Time
}

type ExchangeRate struct {
	Valor string `json:"bid"`
}

// Define as variáveis de timeout e	URL da API de cotação
const MAX_EXCHANGE_API_TIMEOUT = 200 * time.Millisecond
const MAX_DATABASE_TIMEOUT = 10 * time.Millisecond
const EXCHANGE_API_URL = "https://economia.awesomeapi.com.br/json/last/USD-BRL"

// Função main onde é criando o servidor e
// atribuído uma função handler a rota  "/cotacao"
func main() {
	http.HandleFunc("/cotacao", handler)
	http.ListenAndServe(":8080", nil)
}

// Handler para a rota /cotacao
func handler(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()

	log.Println("EXCHANGE_API: Request iniciado")

	// Obtem a cotação
	cotacao, err := GetUsdExchangeRate(ctx)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Println("EXCHANGE_API: Request finalizado")

	// Seta o Header indicando o conteúdo do retorno (JSON)
	w.Header().Set("Content-Type", "application/json")
	// Serializa o JSON e entrega para w
	// w = writer responsável por escrever a response
	// algo semelhante com o código em JS = res.send(usdbrl)
	json.NewEncoder(w).Encode(cotacao)

	log.Println("DATABASE: Insert iniciado")

	// Salva a cotação no banco de dados
	err = Save(ctx, cotacao.Valor)
	if err != nil {
		log.Printf("erro ao tentar salvar cotação: %s", err.Error())
		return
	}

	log.Println("DATABASE: Insert Finalizado")
}

// Função que realiza a requisição para a API e retorna a cotação
func GetUsdExchangeRate(ctx context.Context) (*ExchangeRate, error) {

	ctx, cancel := context.WithTimeout(ctx, MAX_EXCHANGE_API_TIMEOUT)
	defer cancel()

	// Monta o request
	req, err := http.NewRequestWithContext(ctx, "GET", EXCHANGE_API_URL, nil)
	if err != nil {
		return nil, fmt.Errorf("erro ao montar o request: %s", err.Error())
	}

	// Executa o request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar cotação: %s", err.Error())
	}
	defer resp.Body.Close()

	// Utiliza um map por que a estrutura do response é algo como:
	// {
	// 	"USDBRL": {
	// 		"bid": "5.4714"
	// 	}
	// }
	var result map[string]ExchangeRate
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, fmt.Errorf("erro ao decodificar response: %s", err.Error())
	}

	cotacao := result["USDBRL"]
	return &cotacao, nil
}

func Save(ctx context.Context, bid string) error {

	ctx, cancel := context.WithTimeout(ctx, MAX_DATABASE_TIMEOUT)
	defer cancel()

	// Abre a conexão com o banco de dados
	db, err := gorm.Open(sqlite.Open("./cotacoes.db"), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("erro ao conectar no banco: %w", err)
	}

	// Cria a tabela caso não exista
	err = db.WithContext(ctx).AutoMigrate(&Cotacao{})
	if err != nil {
		return fmt.Errorf("erro migrate: %w", err)
	}

	cotacao := Cotacao{
		Valor: bid,
		Data:  time.Now(),
	}

	// Insere a cotação no banco de dados
	err = db.WithContext(ctx).Create(&cotacao).Error
	if err != nil {
		return fmt.Errorf("erro ao inserir cotação: %w", err)
	}

	// Consulta a cotação no banco de dados para verificar se foi inserida
	/*
		err = db.Find(&cotacao).Error
		if err != nil {
			return fmt.Errorf("erro ao consultar cotação: %w", err)
		}

		log.Printf("Cotação Valor: %s", cotacao.Valor)
		log.Printf("Cotação Data: %s", cotacao.Data)
	*/

	return nil
}
