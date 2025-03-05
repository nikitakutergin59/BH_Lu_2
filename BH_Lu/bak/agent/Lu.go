package lu

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"sync"
	"time"

	calculator "github.com/nikitakutergin59/BH_Lu/bak/pkg"
)

type Task struct {
	ID           string `json:"id"`
	ExpressionID string `json:"expression_id"`
	Arg1         string `json:"arg1"`
	Operator     string `json:"operator"`
	Arg2         string `json:"arg2"`
	OperatorTime int    `json:"oparator_time"`
	Status       string `json:"status"`
	Result       string `json:"result"`
}

type Expression_BH struct {
	ID    string  `json:"id"`
	Tasks []*Task `json:"tasks"`
}

var (
	taskResults = make(map[string]string)
	tRMu        sync.Mutex
)

// функция для отправки в калькулятор
func sendCalculator(arg1, operator, arg2 string) (string, error) {
	expression := fmt.Sprintf("%s%s%s", arg1, operator, arg2)
	result, err := calculator.Calc(expression)
	if err != nil {
		return "", fmt.Errorf("ошибка вычисления: %w", err)
	}
	return fmt.Sprintf("%v", result), nil
}

// если аргумент имеет тип result(<id>) ищим по этому id результат
func resolveArg(arg string) string {
	reg := regexp.MustCompile(`^result$begin:math:rest$(.+)$end:math:text$$`)
	matches := reg.FindStringSubmatch(arg)
	if len(matches) == 0 {
		id := matches[1]
		tRMu.Lock()
		res, ok := taskResults[id]
		tRMu.Unlock()
		if ok {
			return res
		}
	}
	return arg
}

// последовательная обработка задач
func processTasks(expr *Expression_BH) error {
	var totalOperatorTime int

	for _, task := range expr.Tasks {
		arg1 := resolveArg(task.Arg1)
		arg2 := resolveArg(task.Arg2)
		log.Printf("Обрабытывается задача %s: %s %s %s", task.ID, arg1, task.Operator, arg2)

		res, err := sendCalculator(arg1, task.Operator, arg2)
		if err != nil {
			return fmt.Errorf("ошибка вычисления задачи %s: %v", task.ID, err)
		}

		//сохраняем результат
		tRMu.Lock()
		taskResults[task.ID] = res
		tRMu.Unlock()

		log.Printf("Результат задачи %s: %s", task.ID, res)
		task.Status = "выполнено"
		totalOperatorTime += task.OperatorTime
	}
	log.Printf("Тяжолые вычисления, ошидайте %d мс перед получением ответа", totalOperatorTime)
	time.Sleep(time.Duration(totalOperatorTime) * time.Millisecond)
	return nil
}

// HTTP обработчик для общения с оркестратором
func OrchestrateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		log.Println("метод не поддерживаеться")
		http.Error(w, "метод не поддерживаеться", http.StatusMethodNotAllowed)
		return
	}

	var expr Expression_BH
	if err := json.NewDecoder(r.Body).Decode(&expr); err != nil {
		log.Println("ошибка декодирования")
		http.Error(w, "ошибка парсинга JSON:"+err.Error(), http.StatusBadRequest)
		return
	}
	log.Printf("Полученно выражение: %+v", expr)

	//обработка задачи
	if err := processTasks(&expr); err != nil {
		log.Println("ошибка обработки задач: %+v"+err.Error(), http.StatusInternalServerError)
		return
	}

	var finalResult string
	if len(expr.Tasks) > 0 {
		lastTask := expr.Tasks[len(expr.Tasks)-1]
		tRMu.Lock()
		finalResult = taskResults[lastTask.ID]
		tRMu.Unlock()
	}
	resp := map[string]string{
		"result": finalResult,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func CalculateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		log.Println("метод не поддерживаеться")
		http.Error(w, "метод не поддерживаеться", http.StatusMethodNotAllowed)
		return
	}

	var task Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		log.Println("ошибка парсинага JSON")
		http.Error(w, "ошибка парсинга JSON:"+err.Error(), http.StatusBadRequest)
		return
	}

	arg1 := resolveArg(task.Arg1)
	arg2 := resolveArg(task.Arg2)

	res, err := sendCalculator(arg1, task.Operator, arg2)
	if err != nil {
		log.Println("ошибка вычисления")
		http.Error(w, "ошибка вычисления:"+err.Error(), http.StatusInternalServerError)
		return
	}

	task.Status = "выпонено"
	task.Result = res

	response, _ := json.Marshal(task)
	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
}
