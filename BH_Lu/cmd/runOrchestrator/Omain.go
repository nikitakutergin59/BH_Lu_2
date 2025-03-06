package main

import (
	"log"
	"net/http"
	"runtime/debug"

	"github.com/gorilla/mux"

	//lu "github.com/nikitakutergin59/BH_Lu/bak/agent"
	"github.com/nikitakutergin59/BH_Lu/bak/orchestrator"
)

func main() {
	calculationTimes := BHhttp.LoadCalculationTime()
	log.Printf("Loaded calculation times: %+v", calculationTimes) // Логируем структуру времен

	md := BHhttp.NewMemoryData()
	calcContext := BHhttp.CalculateContext{
		Times: calculationTimes,
		Md:    md,
	}

	r := mux.NewRouter()

	r.HandleFunc("/calculate", func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("Паника при обработке /calculate: %v\n%s", err, debug.Stack())
				http.Error(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
			}
		}()
		log.Println("Запрос /calculate")
		BHhttp.CalculateHandler(w, r, md, &calcContext)
		log.Println("Запрос /calculate обработан успешно")
	})

	r.HandleFunc("/expressions", func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("Паника при обработке /expressions: %v\n%s", err, debug.Stack())
				http.Error(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
			}
		}()
		log.Println("Запрос /expressions получен")
		BHhttp.GetExpressionsHandler(w, r, md)
		log.Println("Запрос /expressions обработан успешно")
	})

	r.HandleFunc("/expression/{id}", func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("Паника при обработке /expression/{id}: %v\n%s", err, debug.Stack())
				http.Error(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
			}
		}()
		log.Println("Запрос /expression/{id} получен")
		BHhttp.GetExpressionHandler(w, r, md)
		log.Println("Запрос /expression/{id} обработан успешно")
	})

	r.HandleFunc("/task", func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("Паника при обработке запроса /task: %v\n%s", err, debug.Stack())
				http.Error(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
			}
		}()
		log.Println("Запрос /task получен")
		BHhttp.GetTaskHandler(w, r, md)
		log.Println("Запрос /task обработан успешно")
	})

	r.HandleFunc("/result", func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("Паника при обработке запроса /result: %v\n%s", err, debug.Stack())
				http.Error(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
			}
		}()
		log.Println("Запрос /result получен")
		BHhttp.ReceiveResultHandler(w, r, md)
		log.Println("Запрос /result обработан успешно")
	})

	log.Printf("Сервер запущен на порту 8080") // Используем log.Printf
	log.Fatal(http.ListenAndServe(":8080", r))
}
