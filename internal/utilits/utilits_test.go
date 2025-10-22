package utilits

import (
	// Чтобы временно перехватывать вывод в консоль
	// Для вывода в консоль

	// Для работы со строками
	"testing" // Cтандартная библиотека для тестов Go
	"time"    // Для работы с датой и временем

	"github.com/Evgenymoshrage/Go-Log-Processor/internal/model"
)

// ================================================ Тест LogEntryToString ================================================

func TestLogEntryToString(t *testing.T) {
	entry := model.LogEntry{ // Тестовые данные
		Timestamp:    time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		Method:       "GET",
		URL:          "/index",
		StatusCode:   200,
		ResponseTime: 123,
		IP:           "192.168.0.1",
	}

	got := LogEntryToString(entry) // Вызываем тестируемую функцию и сохраняем результат
	// Определяем ожидаемый текст результата
	expected := "[2024-01-15 10:30:00] GET /index (status: 200, response: 123ms, IP: 192.168.0.1)"

	if got != expected { // Сравниваем строки
		t.Errorf("Ожидалось:\n%s\nПолучено:\n%s", expected, got)
	}
}
