package processor

import (
	"encoding/csv" // Для чтения CSV-файлов построчно
	"fmt"          // Для форматирования строк и вывода ошибок
	"os"           // Для открытия файла
	"strconv"      // Для преобразования строк в числа
	"time"

	"github.com/Evgenymoshrage/final/internal/model" // Импортируем структуры LogEntry и Statistics из пакета internal/model
)

// LoadLogs читает CSV-файл и возвращает срез структур LogEntry
func LoadLogs(filePath string) ([]model.LogEntry, error) {
	file, err := os.Open(filePath) // Открытие файла по указанному пути
	if err != nil {                // Обработка ошибки открытия файла
		return nil, fmt.Errorf("Ошибка открытия файла: %v", err)
	}
	defer file.Close() // Откладываем закрытие файла до конца функции

	reader := csv.NewReader(file) // Cоздаём CSV-ридер, который будет построчно считывать данные из файла
	reader.Comma = ','            // Указываем символ-разделитель в файле
	reader.FieldsPerRecord = 6    // Задаём количество колонок в каждой строке

	// Пропускаем заголовок
	_, err = reader.Read()
	if err != nil {
		return nil, fmt.Errorf("Ошибка чтения заголовка: %v", err)
	}

	var logs []model.LogEntry // Создаём пустой срез для хранения всех логов
	for {
		record, err := reader.Read()
		if err != nil {
			if err.Error() == "EOF" { // Если достигнут конец файла — выходим из цикла
				break
			}
			return nil, fmt.Errorf("Ошибка чтения строки: %v", err)
		}

		statusCode, err := strconv.Atoi(record[4]) // Преобразаем статус в число
		if err != nil {
			return nil, fmt.Errorf("Ошибка преобразования status: %v", err)
		}

		respTime, err := strconv.Atoi(record[5]) // Преобразаем время ответа в число
		if err != nil {
			return nil, fmt.Errorf("Ошибка преобразования response_time: %v", err)
		}

		// Создаём экземпляр структуры LogEntry и заполняем его значениями из текущей строки.
		log := model.LogEntry{
			Timestamp:    record[0],
			IP:           record[1],
			Method:       record[2],
			URL:          record[3],
			StatusCode:   statusCode,
			ResponseTime: respTime,
		}

		logs = append(logs, log) // Добавляем структуру в срез
	}

	return logs, nil // Возвращаем готовый список логов и nil
}

// Воркер читает записи логов из канала и "обрабатывает" их.
func Worker(id int, orderChan <-chan model.LogEntry, done chan<- bool) {
	for log := range orderChan {
		fmt.Printf("Воркер %d начал обработку: [%s] %s %s (status: %d)\n",
			id, log.Timestamp, log.Method, log.URL, log.StatusCode)

		// Имитация обработки
		time.Sleep(time.Millisecond * 10)

		fmt.Printf("Воркер %d закончил обработку: [%s]\n", id, log.Timestamp)
	}

	done <- true // сигнал, что воркер завершил работу
}
