package processor

import (
	"context" // Для управления отменой/таймаутом горутин
	"os"      // Для работы с файлами (создание временного CSV)
	"strings" // Для работы со строками
	"testing" // Cтандартная библиотека для тестов Go
	"time"    // Для работы с датой и временем

	"github.com/Evgenymoshrage/Go-Log-Processor/internal/model"
)

// ============================= Вспомогательная функция для создания тестового CSV ===================================

func createTestCSV(t *testing.T, content string) string {
	t.Helper()
	tmpFile := "test_logs.csv"                          // Имя временного файла
	err := os.WriteFile(tmpFile, []byte(content), 0644) // Создаём файл с содержимым CSV
	if err != nil {
		t.Fatalf("Ошибка создания тестового CSV: %v", err)
	}
	return tmpFile // Возвращаем путь к файлу
}

// ================================================ Тест загрузки логов ================================================

func TestLoadLogs(t *testing.T) {
	// Тестовые данные
	csvContent := `timestamp,ip,method,url,status,response_time 
2024-01-15 10:30:00,192.168.0.1,GET,/index,200,123
2024-01-15 10:30:01,192.168.0.2,POST,/login,404,87`

	filePath := createTestCSV(t, csvContent) // Создаём файл с определенным контентом
	defer os.Remove(filePath)                // Отложенно удаляем файл

	logs, err := LoadLogs(filePath) // Проверка что файл корректно прочитан
	if err != nil {
		t.Fatalf("LoadLogs вернул ошибку: %v", err)
	}

	if len(logs) != 2 { // Проверка что количество записей совпадает
		t.Errorf("Ожидалось 2 записи, получили %d", len(logs))
	}

	if logs[0].StatusCode != 200 || logs[1].StatusCode != 404 { // Проверка что статусы каждой записи соответствуют ожидаемым
		t.Errorf("Статусы не соответствуют ожиданиям: %+v", logs)
	}
}

// ================================================ Тест обработки логов ================================================

func TestProcessLogs(t *testing.T) {
	stats := &model.Statistics{ // Создаём объект статистики, который будем передавать в ProcessLogs
		RequestsByIP: make(map[string]int),
	}
	logEntries := []model.LogEntry{ // Тестовые данные
		{IP: "1.1.1.1", StatusCode: 200, ResponseTime: 100, Timestamp: time.Now()},
		{IP: "2.2.2.2", StatusCode: 500, ResponseTime: 200, Timestamp: time.Now()},
	}

	input := make(chan model.LogEntry, len(logEntries)) // Создаём канал input, куда будем отправлять логи для обработки
	for _, l := range logEntries {
		input <- l
	}
	close(input)

	ctx := context.Background() // Создаём контекст, который позволит при необходимости остановить обработку
	output := ProcessLogs(ctx, input, 2, stats)

	var processed []model.LogEntry // Читаем все обработанные логи из канала output
	for l := range output {        // Добавляем их в срез processed
		processed = append(processed, l)
	}

	if len(processed) != len(logEntries) { // Проверяем, что все записи были обработаны
		t.Errorf("Ожидалось обработать %d записей, получили %d", len(logEntries), len(processed))
	}

	if stats.TotalRequests != 2 || stats.ErrorCount != 1 { // Проверяем, что статистика обновилась корректно
		t.Errorf("Статистика не соответствует ожиданиям: %+v", stats)
	}
}

// ================================================ Тест фильтрации логов ===============================================

func TestFilterLogsChannels(t *testing.T) {
	logs := []model.LogEntry{ // Тестовые данные
		{StatusCode: 200}, {StatusCode: 201},
		{StatusCode: 404}, {StatusCode: 500},
	}

	logs2xx, logs4xx, logs5xx := FilterLogs(logs, 0) // Вызываем функцию FilterLogs, которая распределяет логи по каналам

	// Создаём вспомогательную функцию checkChannelLen, чтобы проверять количество элементов в канале
	checkChannelLen := func(ch chan model.LogEntry, expected int, name string) {
		count := 0
		for range ch { // Перебираем канал, считаем все элементы, пока канал не закроется
			count++
		}
		if count != expected { // Если количество элементов count не равно expected, вызываем t.Errorf
			t.Errorf("Ожидалось %d записей в %s, получили %d", expected, name, count)
		}
	}

	// Проверяем каждый канал
	checkChannelLen(logs2xx, 2, "logs2xx")
	checkChannelLen(logs4xx, 1, "logs4xx")
	checkChannelLen(logs5xx, 1, "logs5xx")
}

// ================================================ Тест вывода статистики ================================================

func TestSummaryStatistics(t *testing.T) {
	stats := &model.Statistics{ // Тестовые данные
		RequestsByIP: map[string]int{
			"1.1.1.1": 5,
			"2.2.2.2": 3,
		},
		TotalRequests:   8,
		ErrorCount:      2,
		AverageRespTime: 150,
	}

	result := SummaryStatistics(stats, 2) // Вызываем функцию SummaryStatistics из пакета processor

	// Проверяем, что в строке result есть ключевые элементы
	if !contains(result, "Всего запросов: 8") || !contains(result, "Топ 2 IP") {
		t.Errorf("Результат SummaryStatistics некорректен:\n%s", result)
	}
}

// Вспомогательная функция для поиска подстроки
func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || (len(s) > len(sub) && (strings.Contains(s, sub))))
}

// Проверяем, что строка s хотя бы не короче подстроки sub
// Если строка s полностью совпадает с sub, возвращаем true
// Если s длиннее подстроки (а значит это не полное совпадение),
// проверяем через стандартную функцию strings.Contains, есть ли sub внутри s
