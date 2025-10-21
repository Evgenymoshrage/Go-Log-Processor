package utilits

import (
	"fmt"          // Для вывода в консоль
	"strings"      // Для работы со строками - Repeat
	"unicode/utf8" // Чтобы корректно считать количество символов в UTF-8

	"github.com/Evgenymoshrage/final/internal/model"
)

// ================================================ Реализует интерфейс fmt.Stringer для красивого вывода ================================================

func LogEntryToString(l model.LogEntry) string {
	return fmt.Sprintf("[%s] %s %s (status: %d, response: %dms, IP: %s)",
		l.Timestamp.Format("2006-01-02 15:04:05"),
		l.Method,
		l.URL,
		l.StatusCode,
		l.ResponseTime,
		l.IP,
	)
}

// ================================================ Красивый консольный вывод с подчеркиваниями ================================================
func PrintCentered(title string, width int) {
	line := strings.Repeat("_", width) // Создаём линию из подчёркиваний нужной ширины
	fmt.Println(line)                  // Печатаем верхнюю линию

	padding := (width - utf8.RuneCountInString(title)) / 2 // Считаем количество пробелов слева для центрирования
	if padding < 0 {                                       // Если заголовок длиннее ширины, просто ставим нулевой отступ
		padding = 0
	}

	fmt.Printf("%s%s\n", strings.Repeat(" ", padding), title) // Печатаем сам заголовок с отступом слева
	fmt.Println(line)                                         //  Печатаем нижнюю линию
}
