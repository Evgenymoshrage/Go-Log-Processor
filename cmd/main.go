package main

import (
	"fmt"
	"log"

	"final/internal/model"
	"final/internal/processor"
)

func main() {
	// Загружаем логи
	logs, err := processor.LoadLogs("internal/testdata/logs.csv")
	if err != nil {
		log.Fatalf("Ошибка загрузки логов: %v", err)
	}

	fmt.Printf("Успешно загружено %d записей\n", len(logs))
	for i, l := range logs[:3] { // покажем первые 3
		fmt.Printf("%d: %+v\n", i+1, l)
	}

	// Создаём каналы
	orderChan := make(chan model.LogEntry, len(logs))
	done := make(chan bool)

	// Запускаем 3 воркера
	for i := 1; i <= 3; i++ {
		go processor.Worker(i, orderChan, done)
	}

	// Отправляем все записи в канал
	for _, log := range logs {
		orderChan <- log
	}
	close(orderChan) // Закрываем канал, чтобы воркеры знали, что задач больше нет

	// Ждём, пока все воркеры завершат работу
	for i := 1; i <= 3; i++ {
		<-done
	}

	fmt.Println("Все воркеры завершили работу ✅")
}
