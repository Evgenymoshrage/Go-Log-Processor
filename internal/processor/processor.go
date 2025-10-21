package processor

import (
	"context"      // Для управления таймаутами и отменой задач
	"encoding/csv" // Для чтения CSV-файлов построчно
	"fmt"          // Для форматирования строк и вывода ошибок
	"io"           // Для работы с потоками ввода-вывода
	"os"           // Для открытия файла
	"sort"         // Для сортировки срезов
	"strconv"      // Для преобразования строк в числа

	// Для преобразования int → string
	"sync" // Для синхронизации горутин (WaitGroup)
	"time" // Для работы с датой и временем

	"github.com/Evgenymoshrage/Go-Log-Processor/internal/model" // Импортируем структуры LogEntry и Statistics из пакета internal/model
)

// ================================================  Загрузка логов ================================================

// LoadLogs читает CSV-файл и возвращает срез структур LogEntry
func LoadLogs(filePath string) ([]model.LogEntry, error) {
	file, err := os.Open(filePath) // Открытие файла по указанному пути
	if err != nil {                // Обработка ошибки открытия файла
		return nil, fmt.Errorf("Ошибка открытия файла: %v", err)
	}
	defer file.Close() // Откладываем закрытие файла до конца функции

	reader := csv.NewReader(file) // Cоздаём CSV-ридер, который будет построчно считывать данные из файла
	reader.Comma = ','            // Указываем символ-разделитель в файле

	// Пропускаем заголовок
	_, err = reader.Read()
	if err != nil {
		return nil, fmt.Errorf("Ошибка чтения заголовка: %v", err)
	}

	var logs []model.LogEntry // Создаём пустой срез для хранения всех логов
	for {
		record, err := reader.Read()

		if err != nil {
			if err == io.EOF { // Если достигнут конец файла — выходим из цикла
				break
			}
			return nil, fmt.Errorf("Ошибка чтения строки: %v", err)
		}

		if len(record) != 6 { // Проверяем формат данных
			return nil, fmt.Errorf("Неверное количество полей в строке: %v", record)
		}

		statusCode, err := strconv.Atoi(record[4]) // Преобразуем статус в число
		if err != nil {
			return nil, fmt.Errorf("Ошибка преобразования status: %v", err)
		}

		t, err := time.Parse("2006-01-02 15:04:05", record[0])
		if err != nil {
			return nil, fmt.Errorf("Ошибка парсинга времени: %v", err)
		}

		respTime, err := strconv.Atoi(record[5]) // Преобразуем время ответа в число
		if err != nil {
			return nil, fmt.Errorf("Ошибка преобразования response_time: %v", err)
		}

		// Создаём экземпляр структуры LogEntry и заполняем его значениями из текущей строки.
		log := model.LogEntry{
			Timestamp:    t,
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

// ================================================ Обработка логов ================================================

func ProcessLogs(ctx context.Context, input <-chan model.LogEntry, numWorkers int, stats *model.Statistics) <-chan model.LogEntry {
	output := make(chan model.LogEntry, 100) // Создаём буферезованный выходной канал, куда воркеры будут отправлять обработанные записи
	var wg sync.WaitGroup                    // Создаём WaitGroup, чтобы знать, когда все воркеры закончили работу
	wg.Add(numWorkers)                       // Увеличиваем счётчик на количество воркеров

	for i := 0; i < numWorkers; i++ { // Запускаем параллельные горутины
		go func(workerID int) {
			defer wg.Done()               // Автоматически уменьшит счётчик WaitGroup после завершения воркера
			for logEntry := range input { // Воркер читает каждую запись из входного канала input
				select {
				case <-ctx.Done(): // Проверяем контекст: если пришёл сигнал отмены, воркер завершает работу
					return
				default: // Продолжаем обработку, если отмены нет
					start := time.Now()
					fmt.Printf("[%s] Воркер %d начал обработку: %s\n",
						start.Format("2006-01-02 15:04:05"), workerID, logEntry)

					// Имитация обработки
					time.Sleep(time.Millisecond * 10)

					UpdateStatistics(stats, logEntry) // Обновляем статистику при каждой обработанной записи для func (s *Statistics)

					end := time.Now()
					fmt.Printf("[%s] Воркер %d закончил обработку: %s\n",
						end.Format("2006-01-02 15:04:05"), workerID, logEntry)

					// Отправляем результат в выходной канал
					select {
					case output <- logEntry:
					case <-ctx.Done():
						return
					}
				}
			}
		}(i + 1) // Для нумерации воркеров с 1
	}

	// Создаём отдельную горутину, которая ждёт завершения всех воркеров
	go func() {
		wg.Wait()
		close(output)
	}()

	return output // Возвращаем канал с обработанными логами
}

// ================================================ Фильтрация логов ================================================

func FilterLogs(logs []model.LogEntry, minStatus int) (chan model.LogEntry, chan model.LogEntry, chan model.LogEntry) {
	logs2xx := make(chan model.LogEntry, len(logs))
	logs4xx := make(chan model.LogEntry, len(logs))
	logs5xx := make(chan model.LogEntry, len(logs))

	go func() {
		defer close(logs2xx)
		defer close(logs4xx)
		defer close(logs5xx)

		for _, logEntry := range logs {
			if logEntry.StatusCode < minStatus { // Tсли код статуса меньше minStatus, лог не отправляется ни в один канал
				continue
			}
			switch { // Классифицируем лог по диапазону HTTP-кодов
			case logEntry.StatusCode >= 200 && logEntry.StatusCode < 300:
				logs2xx <- logEntry
			case logEntry.StatusCode >= 400 && logEntry.StatusCode < 500:
				logs4xx <- logEntry
			case logEntry.StatusCode >= 500 && logEntry.StatusCode < 600:
				logs5xx <- logEntry
			}
		}
	}()

	return logs2xx, logs4xx, logs5xx // Возвращаем три канала
}

// ================================================ Вывод статистики ================================================

func UpdateStatistics(s *model.Statistics, log model.LogEntry) {
	s.Mu.Lock()         // Блокирует доступ к статистике, чтобы другие горутины не могли изменять её одновременно
	defer s.Mu.Unlock() // Гарантирует разблокировку после выхода из функции

	s.TotalRequests++        // Увеличиваем общее количество запросов на 1
	s.RequestsByIP[log.IP]++ // Увеличиваем счётчик для IP, с которого пришёл этот запрос

	// 4xx Ошибки клиента
	// 5xx Ошибки сервера
	// Если код ответа сервера — ошибка (4xx или 5xx), увеличиваем количество ошибок
	if log.StatusCode >= 400 {
		s.ErrorCount++
	}

	n := float64(s.TotalRequests) // Чтобы можно было делить числа с плавающей точкой

	// Формула пересчёта среднего без пересуммирования всех данных
	// Новое_среднее = (предыдущее_среднее × (кол-во_старых) + новое_значение) / (новое_кол-во)
	s.AverageRespTime = ((s.AverageRespTime * (n - 1)) + float64(log.ResponseTime)) / n
}

// SummaryStatistics — возвращает красиво отформатированную статистику
// topN — сколько IP показать в топе
func SummaryStatistics(s *model.Statistics, topN int) string {
	s.Mu.Lock()         // Блокируем доступ к статистике, чтобы другие горутины не мешали
	defer s.Mu.Unlock() // Разблокируем после выхода из функции

	// Создаём срез, в который скопируем IP и количество запросов, чтобы потом отсортировать
	topIPs := make([]struct {
		IP    string
		Count int
	}, 0, len(s.RequestsByIP))

	// Проходим по карте всех IP и добавляем их в срез topIPs
	for ip, count := range s.RequestsByIP {
		topIPs = append(topIPs, struct {
			IP    string
			Count int
		}{ip, count})
	}

	// Сортируем topIPs по количеству запросов в порядке убывания
	sort.Slice(topIPs, func(i, j int) bool {
		return topIPs[i].Count > topIPs[j].Count
	})

	// Ограничиваем список, чтобы оставить только топ N IP
	// «Если IP-адресов в списке больше, чем нужно — оставь только первые N, остальные отбрось»
	if len(topIPs) > topN {
		topIPs = topIPs[:topN]
	}

	// Формируем результат в виде строки
	result := fmt.Sprintf(
		"Всего запросов: %d\n"+
			"Ошибок (4xx/5xx): %d\n"+
			"Среднее время ответа: %.2f мс\n"+
			"Топ %d IP:\n",
		s.TotalRequests, s.ErrorCount, s.AverageRespTime, topN,
	)

	// Добавляем построчно информацию о каждом IP из топа
	for i, ip := range topIPs {
		result += fmt.Sprintf("  %d. %s — %d запросов\n", i+1, ip.IP, ip.Count)
	}

	return result // Возвращаем готовую строку со статистикой
}
