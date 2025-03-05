package BHhttp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	//"runtime/debug"
	"strconv"
	"sync"
	"time"

	lu "github.com/nikitakutergin59/BH_Lu/bak/agent"
	tokens "github.com/nikitakutergin59/BH_Lu/bak/tokens"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// хранение выражения
type Expression_BH struct {
	ID         string     `json:"id"`
	Expression string     `json:"expression"`
	Status     string     `json:"status"`
	Result     string     `json:"result"`
	Tasks      []*Task_BH `json:"task_BH"`
	Error      string     `json:"error"`
}

// хранение задачи
type Task_BH struct {
	ID           string        `json:"id"`
	ExpressionID string        `json:"expression_id"`
	Arg_1        string        `json:"arg1"`
	Arg_2        string        `json:"arg_2"`
	Operator     string        `json:"operator"`
	OperatorTime time.Duration `json:"oTime"`
	Result       string        `json:"result"`
	Status       string        `json:"status"`
}

// хранение данных в памяти
type MemoryData struct {
	Expression map[string]*Expression_BH
	Task       map[string]*Task_BH
	mu         sync.RWMutex
}

// конструктор для MemoryData
func NewMemoryData() *MemoryData {
	return &MemoryData{
		Expression: make(map[string]*Expression_BH),
		Task:       make(map[string]*Task_BH),
	}
}

// добавляем выражения в MemoryData
func AddExpression(md *MemoryData, expr *Expression_BH) {
	md.mu.Lock()
	defer md.mu.Unlock()
	md.Expression[expr.ID] = expr
}

// для ответа на запрос задачи
type TaskResponse struct {
	Task *Task_BH `json:"task"`
}

// для ответа на запрос выражения
type ExpressionResponse struct {
	Expression *Expression_BH `json:"expression"`
}

// для ответа на запрос спика вырадений
type ExpressionsResponse struct {
	Expressions []*Expression_BH `json:"expressions"`
}

// для запроса на отправку результата
type ResultRequest struct {
	ID     string `json:"id"`
	Result string `json:"result"`
}

// для запроса на вычисление выражений
type CalculateRequest struct {
	ID         string `json:"id"`
	Expression string `json:"expression"`
}

// для хранения времени выполнения вычислений
type CalculateTime struct {
	Addition       time.Duration
	Subtraction    time.Duration
	Multiplication time.Duration
	Division       time.Duration
}

// загрузка времени из переменных окружения
func LoadCalculationTime() CalculateTime {
	return CalculateTime{
		Addition:       getEnvAsInt("TIME_ADDITION_MS", 100),
		Subtraction:    getEnvAsInt("TIME_SUBTRACTION_MS", 100),
		Multiplication: getEnvAsInt("TIME_MULTIPLICATION_MS", 100),
		Division:       getEnvAsInt("TIME_DIVISION_MS", 100),
	}
}

// вспомогательная функция для получения переменной окружения в виде int
func getEnvAsInt(name string, defaultValue int) time.Duration {
	valueStr := os.Getenv(name)
	if valueStr == "" {
		return time.Duration(defaultValue) * time.Millisecond
	}
	valueInt, err := strconv.Atoi(valueStr)
	if err != nil {
		log.Printf("Ошибка при чтении переменной окружения %s: %v, используется значение по умолчанию %d", name, err, defaultValue)
		return time.Duration(defaultValue) * time.Millisecond
	}
	return time.Duration(valueInt) * time.Millisecond
}

// для передачи контекста вычисления
type CalculateContext struct {
	Times CalculateTime
	Md    *MemoryData
}

// присвоение к операторам переменных
func (c *CalculateContext) GetOperatorTime(operator string) (time.Duration, error) {
	switch operator {
	case "+":
		return c.Times.Addition, nil
	case "-":
		return c.Times.Subtraction, nil
	case "*":
		return c.Times.Multiplication, nil
	case "/":
		return c.Times.Division, nil
	default:
		return 0, fmt.Errorf("неизвестный оператор: %s", operator)
	}
}

// генерируй ID для каждого выражения
func generateID() (string, error) {
	uuid, err := uuid.NewRandom()
	if err != nil {
		return "", fmt.Errorf("ошибка при генерации UUID: %w", err)
	}
	return uuid.String(), nil
}

func CalculateHandler(w http.ResponseWriter, r *http.Request, md *MemoryData, calcContext *CalculateContext) {
	log.Println("CalculateHandler: начало обработки запроса") // 1

	if r.Method != http.MethodPost {
		log.Println("CalculateHandler: метод не POST") // 2
		http.Error(w, "метод не разрешон", http.StatusMethodNotAllowed)
		return
	}
	log.Println("CalculateHandler: метод POST") // 3

	var req CalculateRequest
	log.Println("CalculateHandler: декодируем JSON") // 4
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("CalculateHandler: ошибка декодирования JSON: %v", err) // 5
		http.Error(w, "неверный запрос", http.StatusBadRequest)
		return
	}
	log.Printf("CalculateHandler: декодирован JSON: %+v", req) // 6

	id, err := generateID()
	log.Println("CalculateHandler: генерируем ID") // 7
	if err != nil {
		log.Printf("CalculateHandler: ошибка генерации ID: %v", err) // 8
		http.Error(w, "ошибка генерации ID", http.StatusInternalServerError)
		return
	}
	log.Printf("CalculateHandler: сгенерирован ID: %s", id) // 9

	expr := &Expression_BH{
		ID:         id,
		Expression: req.Expression,
		Status:     "получено",
		Tasks:      []*Task_BH{},
	}
	log.Printf("CalculateHandler: создано выражение: %+v", expr) // 10

	md.mu.RLock()
	log.Println("CalculateHandler: RLock для md.Expression") // 11
	md.Expression[id] = expr
	md.mu.RUnlock()
	log.Println("CalculateHandler: RUnlock для md.Expression") // 12

	AddExpression(md, expr)
	log.Println("CalculateHandler: добавлено выражение в md") // 13

	tokens, err := tokens.Tokenize_BH(req.Expression)
	log.Println("CalculateHandler: токенизируем выражение") // 14
	if err != nil {
		log.Printf("CalculateHandler: ошибка токенизации: %v", err) // 15
		expr.Status = "error"
		expr.Error = err.Error()
		http.Error(w, "ошибка токенизации:", http.StatusBadRequest)
		return
	}
	log.Printf("CalculateHandler: токены: %+v", tokens) // 16

	errChan := make(chan error, 1)
	log.Println("CalculateHandler: создаем канал ошибок") // 17

	go func() {
		log.Println("CalculateHandler: запускаем горутину processExpression") // 18
		err := processExpression(expr, tokens, calcContext)
		if err != nil {
			log.Printf("CalculateHandler: ошибка в processExpression: %v", err) // 19
			md.mu.Lock()
			expr.Status = "error"
			expr.Error = err.Error()
			md.mu.Unlock()
			errChan <- err
			return
		}
		log.Println("CalculateHandler: processExpression завершена успешно") // 20
		errChan <- nil
	}()

	log.Println("CalculateHandler: ожидаем завершения горутины") // 21
	err = <-errChan
	log.Println("CalculateHandler: горутина завершена") // 22
	if err != nil {
		log.Printf("CalculateHandler: ошибка из горутины: %v", err) // 23
		return
	}

	// создаём горутину для отправки запроса с целью получить выражение по id
	var wg sync.WaitGroup
	wg.Add(1)

	body_str_chan := make(chan string, 1)
	defer close(body_str_chan)
	go func(id string, expression Expression_BH) {
		defer wg.Done()
		getUrl := fmt.Sprintf("http://localhost:8080/expression/%s", expression.ID)

		resp, err := http.Get(getUrl)
		if err != nil {
			log.Printf("ошибка при отправке GET запроса: %v", err)
			http.Error(w, "ошибка отправки запроса", http.StatusInternalServerError)
			body_str_chan <- ""
			return
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Printf("ошибка при чтение тела запроса: %v", err)
			http.Error(w, "не найденно тело отвена", http.StatusNotFound)
			body_str_chan <- ""
			return
		}
		str_body := string(body)
		body_str_chan <- str_body
	}(id, *expr)

	wg.Wait()
	select {
	case otvet := <-body_str_chan:
		if otvet == "" {
			log.Println("Горутина вернула ошибку")
			http.Error(w, "ошибка при получении выражения", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{"id": id, "под задачи": otvet})
		log.Println("CalculateHandler: запрос обработан успешно") // 24
	default:
		log.Println("Горутина вернула ошибку")
	}

}

// обработчик на получения списка выражений
func GetExpressionsHandler(w http.ResponseWriter, r *http.Request, md *MemoryData) {
	md.mu.RLock()
	defer md.mu.RUnlock()

	expressions := make([]*Expression_BH, 0, len(md.Expression))
	for _, expr := range md.Expression {
		expressions = append(expressions, expr)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ExpressionsResponse{Expressions: expressions})
}

// Обрабатываем выражение, создавая подзадачи для вычисления
func processExpression(expr *Expression_BH, tokens_BH []tokens.Token, calcContext *CalculateContext) error {
	log.Println("processExpression: начало") // 1

	var (
		stack     [][]tokens.Token
		newTokens []tokens.Token
	)

	for _, token := range tokens_BH {
		if token.Type == tokens.TOKEN_PARENT_OPEN {
			stack = append(stack, []tokens.Token{}) // Добавляем новый уровень стека
		} else if token.Type == tokens.TOKEN_PARENT_CLOSE {
			if len(stack) == 0 {
				return fmt.Errorf("некорректные скобки в выражении")
			}

			lastExpr := stack[len(stack)-1] // Достаем последнее выражение
			stack = stack[:len(stack)-1]    // Убираем его из стека

			// Если внутри пусто или меньше 3 токенов — ошибка
			if len(lastExpr) < 3 {
				return fmt.Errorf("некорректное выражение внутри скобок: %v", lastExpr)
			}

			// Создаём новую задачу для выражения внутри скобок
			subExprID, err := generateID()
			if err != nil {
				log.Printf("processExpression: ошибка генерации ID: %v", err)
				return fmt.Errorf("ошибка генерации ID: %w", err)
			}

			log.Printf("processExpression: subExprID = %s", subExprID)
			log.Printf("processExpression: создана задача %s для %v", subExprID, lastExpr)

			// Рекурсивно разбираем выражение внутри скобок
			subExpr := &Expression_BH{
				ID:    subExprID,
				Tasks: []*Task_BH{},
			}

			err = processExpression(subExpr, lastExpr, calcContext)
			if err != nil {
				return err
			}

			// Создаём subTask для вычисления результата подвыражения
			subTask := &Task_BH{
				ID:           subExprID,
				ExpressionID: expr.ID,
				Arg_1:        "result(" + subExprID + ")",
				Operator:     "",
				Arg_2:        "",
				Status:       "ожидание",
			}

			// Добавляем подзадачу в контекст
			calcContext.Md.mu.Lock()
			calcContext.Md.Task[subExprID] = subTask
			expr.Tasks = append(expr.Tasks, subTask)
			calcContext.Md.mu.Unlock()

			// Заменяем выражение в скобках на result(ID)
			newTokens = append(newTokens, tokens.Token{
				Type:  tokens.TOKEN_NUMBER,
				Value: fmt.Sprintf("result(%s)", subExprID),
			})

		} else {
			if len(stack) > 0 {
				// Если есть открытая скобка, добавляем токен внутрь неё
				stack[len(stack)-1] = append(stack[len(stack)-1], token)
			} else {
				newTokens = append(newTokens, token) // Вне скобок — просто добавляем в новый список
			}
		}
	}

	// После разбора скобок, обрабатываем оставшиеся операторы
	if len(newTokens) >= 3 {
		err := splitExpression(expr, newTokens, calcContext)
		if err != nil {
			return err
		}
	}

	log.Println("processExpression: завершение") // 12
	return nil
}

func splitExpression(expr *Expression_BH, tokens []tokens.Token, calcContext *CalculateContext) error {
	if len(tokens) < 3 {
		return fmt.Errorf("недостаточно токенов для выражения: %v", tokens)
	}

	for i := 0; i < len(tokens)-2; i += 2 {
		arg1 := tokens[i].Value
		operator := tokens[i+1].Value
		arg2 := tokens[i+2].Value

		taskID, err := generateID()
		if err != nil {
			return fmt.Errorf("ошибка генерации ID: %w", err)
		}

		operatorTime, err := calcContext.GetOperatorTime(operator)
		if err != nil {
			return fmt.Errorf("ошибка получения времени оператора %w", err)
		}

		task := &Task_BH{
			ID:           taskID,
			ExpressionID: expr.ID,
			Arg_1:        arg1,
			Operator:     operator,
			Arg_2:        arg2,
			OperatorTime: operatorTime,
			Status:       "получено",
		}

		// Добавляем задачу в контекст
		calcContext.Md.mu.Lock()
		calcContext.Md.Task[taskID] = task
		expr.Tasks = append(expr.Tasks, task)
		calcContext.Md.mu.Unlock()
	}

	return nil
}

// обработчик на получения выражения по ID
func GetExpressionHandler(w http.ResponseWriter, r *http.Request, md *MemoryData) {
	vars := mux.Vars(r)
	id := vars["id"]

	md.mu.RLock()
	expr, ok := md.Expression[id]
	md.mu.RUnlock()

	if !ok {
		http.Error(w, "выражение не найдено", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ExpressionResponse{Expression: expr})
}

// обработчик на получение задачи
func GetTaskHandler(w http.ResponseWriter, r *http.Request, md *MemoryData) {
	md.mu.Lock()
	defer md.mu.Unlock()

	for _, task := range md.Task {
		if task.Status == "получено" {
			task.Status = "выполняеться"
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(TaskResponse{Task: task})
			return
		}
	}

	http.Error(w, "задача не найдена", http.StatusNotFound)
}

// обработчик на получение результата
//func ReceiveResultHandler(w http.ResponseWriter, r *http.Request, md *MemoryData) {
//	var req ResultRequest
//	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
//		http.Error(w, "неверный запрос", http.StatusBadRequest)
//		return
//	}
//
//	md.mu.Lock()
//	defer md.mu.Unlock()
//
//	task, ok := md.Task[req.ID]
//	if !ok {
//		http.Error(w, "задача не найденна", http.StatusNotFound)
//		return
//	}
//
//	task.Result = req.Result
//	task.Status = "выпонено"
//
//	w.WriteHeader(http.StatusOK)
//}

// для отправки задачи агенту
var mu sync.Mutex

func SendTaskAgent(task *lu.Task) (string, error) {
	log.Printf("отправляем агенту %+v", task)
	agentURL := "http://localhost:8081/calculate"

	taskData, err := json.Marshal(task)
	if err != nil {
		return "", fmt.Errorf("ошибка кодирования JSON: %v", err)
	}

	resp, err := http.Post(agentURL, "application/json", bytes.NewBuffer(taskData))
	if err != nil {
		return "", fmt.Errorf("ошибка отправки запроса агенту: %v", err)
	}
	defer resp.Body.Close()

	//body, err := io.ReadAll(resp.Body)
	//if err != nil {
	//	return "", fmt.Errorf("ошибка чтения ответа агента: %v", err)
	//}

	var response lu.Task
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("ошибка парсинга ответа: %w", err)
	}

	return response.Result, nil
}

// HTTP обработчик для получения результата от агента
func ReseiveTaskResultHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		log.Println("метод не поддерживаеться")
		http.Error(w, "метод не поддерживаеться", http.StatusMethodNotAllowed)
		return
	}

	var result struct {
		ID     string `json:"id"`
		Result string `json:"result"`
	}
	log.Println("РЕЗУЛЬТАТ", result)

	if err := json.NewDecoder(r.Body).Decode(&result); err != nil {
		log.Println("ошибка парсинга JSON")
		http.Error(w, "ошибка парсинга JSON:"+err.Error(), http.StatusBadRequest)
		return
	}

	mu.Lock()
	log.Printf("Получен результат для задачи %s: %s", result.ID, result.Result)
	mu.Unlock()

	w.WriteHeader(http.StatusOK)
}
