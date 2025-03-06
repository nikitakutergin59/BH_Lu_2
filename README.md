# BH_Lu_2

#Описание
Распределительно вычислительная машина основной функцией которой явзяеться файл orchestrator.go в нём принимаеться выражение и дальше разбиваеться на задачи и отправляеться в агента
- `cmd/` - директория с файлами которые отвечают за запуск серверов
 
  - `runDemon/` - директория с main для агента
    
  - `runOrchestrator/` - директория с main для оркестратора
    
- `bak/` - директория где храниться бек энд часть

  - `agent` - директория где храниться агент
  
  - `orchestrator` - директория оркестратор
  
  - `pkg` - директория с калькулятором
  
  - `tokens` - токенизатор для сервера

#Запуск
1. Установите [Go] (https://go.dev/dl/).
2. Откройте консоль
Win + R
3. Склонируйте проек с GitHub
    ```bash
    git clone https://github.com/nikitakutergin59/BH_Lu_2
    ```
4. Перейдите в директорию с Omain.go
```bash
  cd BH_Lu_2/BH_Lu/cmd/runOrchestrator
```
5. Запустить оркестратор
```bash
  go run Omain.go
```
6. Откройте второе окно командной строки
Win + R
7. Перейдите в директорию с Dmain.go
```bash
 cd BH_Lu_2/BH_Lu/cmd/runDemon
```
8. Запустите агента
```bash
  go run Dmain.go
```

#Взаимодействие

Что бы задать задачу

Пример

```bash
curl -X POST -H "Content-Type: application/json" -d "{\"expression\": \"2 + 3 * 4\"}" http://localhost:8080/calculate
```
вы получите результат

Что бы запросить список всех задач

```bash
curl -X GET http://localhost:8080/expressions
```

Что бы запросить подзадачу
```bash
curl -X GET http://localhost:8080/task
```

Читая этот README, вы держите в руках моё будущее. Для меня IT за последний год стало не просто важной частью жизни, оно стало её смыслом. Я понимаю, что пока не являюсь лучшим программистом, но очень хочу пройти на следующий спринт и готов стараться, чтобы достичь лучших результатов. Да, BH_Lu_2 работает не совсем так, как планировалось, прошу оценивать не слишком строго. Заранее благодарю за понимание.
