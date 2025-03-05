package main

import (
	"log"
	"net/http"

	"github.com/nikitakutergin59/BH_Lu/bak/agent"
)

func main() {
	http.HandleFunc("/calculate", lu.CalculateHandler)

	log.Println("Демон слушает порт 8081...")
	log.Fatal(http.ListenAndServe(":8081", nil))
}
