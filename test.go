// Цель работы: Знакомство с межсервисным взаимодействием и асинхронностью
// Порядок показа: вызвать через insomnia http-метод асинхронного сервиса,
// показать что в основном приложении появился результат, потом вызвать метод основного сервиса напрямую,
// чтобы изменить результат
// Контрольные вопросы: grpc, асинхронность, веб-сервис
// Задание: Создание асинхронного сервиса для отложенного действия (вычисление, моделирование, оплата и тд)
// с одним http-методом для выполнения отложенного действия в вашей системе
// (вычисление, моделирование, оплата и тд). Действие выполняется с задержкой 5-10 секунд,
// результат сервиса случайный, например успех/неуспех, достаточно в результате обновить одно поле в заявке.
// В исходном веб-сервисе также необходимо добавить http-метод для внесения результатов.
// Асинхронный сервис взаимодействует с основным через http, без прямого обращения в БД.
// Добавить псевдо авторизацию в методе основного сервиса - передавать как константу какой-нибудь ключ,
// например на 8 байт, и через if просто проверять на совпадение это поле.
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"
)

const secretKey = "aB3dE4gH"

type ReviewRequest struct {
	Id_request int `json:"id_request"`
	UserID     int `json:"id_user"`
}

type ReviewResult struct {
	Result bool `json:"result"` // Changed to a boolean type
}

type Response struct {
	Message string `json:"message"`
}

func main() {
	http.HandleFunc("/asyncProcess", handleReview)
	log.Println("Сервер был запущен на порту 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleReview(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req ReviewRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		response := Response{
			Message: "Запрос не выполнен",
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	w.WriteHeader(http.StatusOK)
	response := Response{
		Message: "Запрос выполнен",
	}
	json.NewEncoder(w).Encode(response)

	go processReview(req.Id_request, req.UserID)
}

func processReview(id_request int, id_user int) {
	// Имитация задержки выполнения действия
	delay := rand.Intn(6) + 5
	log.Printf("Request ID %d и user ID %d выполняется с задержкой %d секунд", id_request, id_user, delay)
	time.Sleep(time.Duration(delay) * time.Second)

	rand.Seed(time.Now().UnixNano()) // Инициализируем генератор случайных чисел

	result := rand.Intn(2) == 1 // Randomly decide true or false

	log.Printf("Request ID %d и user ID %d завершился с результатом %t", id_request, id_user, result)

	sendResult(id_request, id_user, result)
}

func sendResult(id_request int, id_user int, result bool) error {
	// Создаем структуру данных для отправки
	reviewResult := struct {
		Result    bool   `json:"result"`
		UserID    int    `json:"id_user"`
		SecretKey string `json:"secretKey"`
	}{
		Result:    result,
		UserID:    id_user,
		SecretKey: secretKey,
	}

	// Преобразуем структуру данных в JSON
	jsonData, err := json.Marshal(reviewResult)
	if err != nil {
		return fmt.Errorf("Ошибка при маршалинге JSON данных: %v", err)
	}

	// Создаем запрос на PUT-запрос
	req, err := http.NewRequest("PUT", fmt.Sprintf("http://0.0.0.0:8000/api/update-request-status/%d/", id_request), bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("Ошибка при создании PUT-запроса: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("Ошибка при выполнении PUT-запроса: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Ошибка при отправке результата: код состояния %d", resp.StatusCode)
	}

	log.Printf("Отправлено Request ID [%d] - Результат: %t", id_request, result)

	return nil
}
