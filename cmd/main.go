package main

import (
	"context" // Для управления таймаутами и отменой задач
	"fmt"     // Для форматирования строк и вывода ошибок
	"log"     // Для логирования сообщений
	"time"    // Для работы с датой и временем

	"github.com/Evgenymoshrage/final/internal/model"
	"github.com/Evgenymoshrage/final/internal/processor"
	"github.com/Evgenymoshrage/final/internal/utilits"
)

func main() {

	// ================================================  Загрузка логов ================================================

	utilits.PrintCentered("Загружаем логи!", 120)
	// Загружаем логи
	logs, err := processor.LoadLogs("internal/testdata/logs.csv")
	if err != nil {
		log.Fatalf("Ошибка загрузки логов: %v", err)
	}

	fmt.Printf("Успешно загружено %d записей\n", len(logs))
	for i, l := range logs[:2] { // покажем первые 2
		fmt.Printf("%d: %+v\n", i+1, utilits.LogEntryToString(l))
	}
	fmt.Println("...")
	for i, l := range logs[len(logs)-2:] { // покажем последние  2
		fmt.Printf("%d: %+v\n", len(logs)-2+i+1, utilits.LogEntryToString(l))
	}

	// ================================================ Обработка логов ================================================
	utilits.PrintCentered("Воркеры начинают работу!", 120)
	// Общий таймаут на всю систему — 10 секунд
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	numWorkers := 5
	inputChan := make(chan model.LogEntry) // Канал для воркеров
	stats := &model.Statistics{            // Создаём объект статистики
		RequestsByIP: make(map[string]int),
	}

	outputChan := processor.ProcessLogs(ctx, inputChan, numWorkers, stats)

	go func() {
		for _, logEntry := range logs {
			inputChan <- logEntry
		}
		close(inputChan) // Закрываем канал, чтобы воркеры знали, что задач больше нет
	}()
	var processedLogs []model.LogEntry
	for log := range outputChan {
		processedLogs = append(processedLogs, log)
	}

	// Сообщение о завершении всех воркеров
	utilits.PrintCentered("Все воркеры завершили работу!", 120)

	// ================================================ Фильтрация логов ================================================

	// Фильтруем уже после завершения воркеров
	logs2xx, logs4xx, logs5xx := processor.FilterLogs(processedLogs, 200)
	utilits.PrintCentered("Запускается фильтрация!", 120)
	fmt.Println("=== 2xx ===")
	for log := range logs2xx {
		fmt.Printf("%d %s\n", log.StatusCode, log.URL)
	}

	fmt.Println("=== 4xx ===")
	for log := range logs4xx {
		fmt.Printf("%d %s\n", log.StatusCode, log.URL)
	}

	fmt.Println("=== 5xx ===")
	for log := range logs5xx {
		fmt.Printf("%d %s\n", log.StatusCode, log.URL)
	}
	utilits.PrintCentered("Фильтрация окончена!", 120)

	// ================================================ Вывод статистики ================================================
	utilits.PrintCentered("Статистика:", 120)
	fmt.Println(processor.SummaryStatistics(stats, 5)) // Печатаем статистику
}
