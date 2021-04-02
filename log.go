// Package log расширяет стандартный go логгер для вывода отладочной
// информации о ходе работы приложения, разделяя его на несколько
// уровней важности.
package log

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	acolor "github.com/VolkovRA/GoAColor"
)

// Level описывает уровень важности логируемых сообщений.
//
// Необходим для разделения сообщений журнала по уровню важности.
// Все доступные значения и их описание для применения перечислены в
// соответствующих константах.
//
// Возможные значения:
//
// - TRACE - Журналы, содержащие наиболее подробные сообщения.
// Эти сообщения могут содержать конфиденциальные данные приложения.
// Эти сообщения по умолчанию отключены, и их никогда не следует включать
// в рабочей среде.
//
// - DEBUG - Журналы, используемые для интерактивного исследования во
// время разработки. Эти журналы в основном содержат сведения, полезные
// при отладки и не представляющие ценности в долгосрочной перспективе.
//
// - INFO - Журналы, отслеживающие общий поток работы приложения.
// Эти журналы должны быть полезны в долгосрочной перспективе.
//
// - WARN - Журналы, которые показывают ненормальное или неожиданное
// событие в потоке приложения, но не вызывают прекращение выполнения
// приложения каким-либо образом.
//
// - ERROR - Журналы, описывающие неустранимый сбой приложения или системы
// либо неустранимый сбой, который требует немедленного внимания.
type Level int32

// Уровни важности логируемых сообщений.
// Тут перечислены все доступные уровни важности и их описание для применения.
const (

	// TRACE - Журналы, содержащие наиболее подробные сообщения.
	// Эти сообщения могут содержать конфиденциальные данные приложения.
	// Эти сообщения по умолчанию отключены, и их никогда не следует включать
	// в рабочей среде.
	TRACE Level = iota

	// DEBUG - Журналы, используемые для интерактивного исследования во
	// время разработки. Эти журналы в основном содержат сведения, полезные
	// при отладки и не представляющие ценности в долгосрочной перспективе.
	// Используется по умолчанию.
	DEBUG

	// INFO - Журналы, отслеживающие общий поток работы приложения.
	// Эти журналы должны быть полезны в долгосрочной перспективе.
	INFO

	// WARN - Журналы, которые показывают ненормальное или неожиданное
	// событие в потоке приложения, но не вызывают прекращение выполнения
	// приложения каким-либо образом.
	WARN

	// ERROR - Журналы, описывающие неустранимый сбой приложения или системы
	// либо неустранимый сбой, который требует немедленного внимания.
	ERROR
)

// Дефолтный логгер.
var std = New(os.Stderr, DEBUG)

// Logger описывает один экземпляр логгера.
//
// По умолчанию используется дефолтный экземпляр логгера, ссылку на который
// Вы можете получить с помощью: log.Default(). Он нацелен на стандартный поток
// вывода сообщений об ошибках. Вы также можете создать собственный экземпляр
// логгера и нацелить его на произвольный поток вывода с помощью конструктора:
// log.New()
type Logger struct {

	// Цветной текст.
	//
	// Если задано true, к тексту будет применяться раскраска с помощью
	// управляющих ANSI символов.
	//
	// По умолчанию: true
	Color bool

	// Время в UTC.
	//
	// Если true, логгер будет использовать нулевой часовой пояс, установленный
	// в локальной системе.
	//
	// По умолчанию: true.
	UTC bool

	// Отображение заголовка. (Целиком)
	//
	// Если true, логгер добавляет в каждое сообщение заголовок с системной
	// информацией: время, уровень важности и т.п.
	//
	// По умолчанию: true.
	Head bool

	// Отображение уровня важности в заголовке.
	//
	// Если true, в заголовке каждого сообщения будет присутствовать маркер
	// уровня важности данного сообщения: [LEVEL].
	//
	// По умолчанию: true.
	HeadLevel bool

	// Отображение даты в заголовке.
	//
	// Если true, в заголовке каждого сообщения будет присутствовать дата: DD.MM.YYYY.
	//
	// По умолчанию: true.
	HeadDate bool

	// Отображение времени в заголовке.
	//
	// Если true, в заголовке каждого сообщения будет присутствовать время: HH:MM:SS.
	//
	// По умолчанию: true.
	HeadTime bool

	// Отображение микросекунд в заголовке. (Работает только при включенном HeadTime)
	//
	// Если true, в заголовке каждого сообщения будут присутствовать микросекунды: HH:MM:SS.000000
	//
	// По умолчанию: false.
	HeadMC bool

	mu    sync.Mutex // Атомарная запись.
	out   io.Writer  // Назначение для вывода сообщений.
	level Level      // Уровень логируемых сообщений.
	buf   []byte     // Буфер для сложения текста при записи.
}

// New создаёт новый логгер.
// Вы можете указать цель назначения всех сообщений журнала.
func New(out io.Writer, level Level) *Logger {
	return &Logger{
		out:       out,
		level:     level,
		Color:     true,
		UTC:       true,
		Head:      true,
		HeadLevel: true,
		HeadDate:  true,
		HeadTime:  true,
		HeadMC:    false,
	}
}

// Default дефолтный логгер, используемый по умолчанию.
// Вы можете создать собственный логгер, используя вызов: log.New().
func Default() *Logger {
	return std
}

// Записать заголовки сообщения.
func (l *Logger) writeHeader(buf *[]byte, level Level) {

	// Метка уровня:
	if l.HeadLevel {
		*buf = append(*buf, l.getHeaderLevel(level)...)
	}

	// Цвет заголовка:
	if l.Color {
		if level == ERROR {
			*buf = append(*buf, acolor.Apply(acolor.Red)...)
		} else {
			*buf = append(*buf, acolor.Apply(acolor.BlackHi)...)
		}
	}

	// Заголовки:
	if l.HeadDate || l.HeadTime {
		var now = time.Now()
		if l.UTC {
			now = now.UTC()
		}
		if l.HeadDate {
			year, month, day := now.Date()

			itoa(buf, day, 2)
			*buf = append(*buf, '.')
			itoa(buf, int(month), 2)
			*buf = append(*buf, '.')
			itoa(buf, year, 4)
			*buf = append(*buf, ' ')
		}
		if l.HeadTime {
			hour, min, sec := now.Clock()
			itoa(buf, hour, 2)
			*buf = append(*buf, ':')
			itoa(buf, min, 2)
			*buf = append(*buf, ':')
			itoa(buf, sec, 2)

			if l.HeadMC {
				*buf = append(*buf, '.')
				itoa(buf, now.Nanosecond()/1000, 6)
			}

			*buf = append(*buf, ' ')
		}
	}

	// Конец заголовка:
	var length = len(*buf)
	if length == 0 {
		return
	}

	*buf = (*buf)[0 : length-1]

	if l.Color {
		*buf = append(*buf, (": " + acolor.Clear())...)
	} else {
		*buf = append(*buf, ": "...)
	}
}

// Получить метку уровня логирования.
func (l *Logger) getHeaderLevel(level Level) string {
	if l.Color {
		switch level {
		case INFO:
			return acolor.Apply(acolor.Bold, acolor.Green) + "[INFO]  " + acolor.Clear()
		case WARN:
			return acolor.Apply(acolor.Bold, acolor.Yellow) + "[WARN]  " + acolor.Clear()
		case TRACE:
			return acolor.Apply(acolor.Bold, acolor.White) + "[TRACE] " + acolor.Clear()
		case DEBUG:
			return acolor.Apply(acolor.Bold, acolor.Cyan) + "[DEBUG] " + acolor.Clear()
		default:
			return acolor.Apply(acolor.Bold, acolor.Red) + "[ERROR] " + acolor.Clear()
		}
	} else {
		switch level {
		case INFO:
			return "[INFO]  "
		case WARN:
			return "[WARN]  "
		case TRACE:
			return "[TRACE] "
		case DEBUG:
			return "[DEBUG] "
		default:
			return "[ERROR] "
		}
	}
}

// Запись инта в строку с фиксированной длиной.
func itoa(buf *[]byte, i int, wid int) {
	var b [20]byte
	bp := len(b) - 1
	for i >= 10 || wid > 1 {
		wid--
		q := i / 10
		b[bp] = byte('0' + i - q*10)
		bp--
		i = q
	}
	b[bp] = byte('0' + i)
	*buf = append(*buf, b[bp:]...)
}

// Записать сообщение в журнал.
func (l *Logger) write(level Level, v ...interface{}) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Шапка:
	l.buf = l.buf[:0]
	if l.Head {
		l.writeHeader(&l.buf, level)
	}

	// Тело:
	if l.Color && level == ERROR {
		l.buf = append(l.buf, (acolor.Apply(acolor.Red) + fmt.Sprint(v...) + acolor.Clear() + "\n")...)
	} else {
		l.buf = append(l.buf, (fmt.Sprint(v...) + "\n")...)
	}

	// Вывод:
	_, err := l.out.Write(l.buf)

	return err
}

// Level указывает текущий уровень важности логируемых сообщений.
//
// Если сообщение не соответствует уровню важности, оно не попадает в журнал.
//
// По умолчанию: LevelTrace. (В журнал попадают все сообщения)
func (l *Logger) Level() Level {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.level
}

// SetLevel устанавливает уровень важности логируемых сообщений.
// Доступные значения Level смотрите в константах пакета.
func (l *Logger) SetLevel(level Level) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

// Output цель вывода сообщений лога.
// По умолчанию: os.Stderr.
func (l *Logger) Output() io.Writer {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.out
}

// SetOutput устанавливает цель вывода сообщений журнала.
func (l *Logger) SetOutput(w io.Writer) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.out = w
}

// IsLevel проверяет актуальность указанного уровня логирования.
//
// Это полезно, если вам нужно проверить, выводится для в данный
// момент указанный уровень логируемых сообщений. Например, перед
// выполнение дорогой операции для создания сообщения для лога.
//
// Возвращает true, если указанный уровень логирования актуален.
func (l *Logger) IsLevel(level Level) bool {
	return level >= l.level
}

// IsError проверяет актуальность уровня логирования: ERROR.
// Возвращает true, если сообщения этого уровня пишутся в журнал.
func (l *Logger) IsError() bool {
	return l.IsLevel(ERROR)
}

// IsWarn проверяет актуальность уровня логирования: WARN.
// Возвращает true, если сообщения этого уровня пишутся в журнал.
func (l *Logger) IsWarn() bool {
	return l.IsLevel(WARN)
}

// IsInfo проверяет актуальность уровня логирования: INFO.
// Возвращает true, если сообщения этого уровня пишутся в журнал.
func (l *Logger) IsInfo() bool {
	return l.IsLevel(INFO)
}

// IsDebug проверяет актуальность уровня логирования: DEBUG.
// Возвращает true, если сообщения этого уровня пишутся в журнал.
func (l *Logger) IsDebug() bool {
	return l.IsLevel(DEBUG)
}

// IsTrace проверяет актуальность уровня логирования: TRACE.
// Возвращает true, если сообщения этого уровня пишутся в журнал.
func (l *Logger) IsTrace() bool {
	return l.IsLevel(TRACE)
}

// Error выводит сообщение об ошибке и завершает работу приложения.
// Пишет сообщение о фатальной ошибке и вызывает: os.Exit(1).
func (l *Logger) Error(v ...interface{}) {
	if ERROR < l.level {
		return
	}

	l.write(ERROR, v...)
	os.Exit(1)
}

// Warn выводит предупреждение.
// Вызов игнорируется, если уровень важности логируемых сообщений не соответствует: WARN.
func (l *Logger) Warn(v ...interface{}) {
	if WARN < l.level {
		return
	}

	l.write(WARN, v...)
}

// Info выводит информационное сообщение.
// Вызов игнорируется, если уровень важности логируемых сообщений не соответствует: INFO.
func (l *Logger) Info(v ...interface{}) {
	if INFO < l.level {
		return
	}

	l.write(INFO, v...)
}

// Debug выводит отладочное сообщение.
// Вызов игнорируется, если уровень важности логируемых сообщений не соответствует: DEBUG.
func (l *Logger) Debug(v ...interface{}) {
	if DEBUG < l.level {
		return
	}

	l.write(DEBUG, v...)
}

// Trace выводит произвольное сообщение.
// Вызов игнорируется, если уровень важности логируемых сообщений не соответствует: TRACE.
func (l *Logger) Trace(v ...interface{}) {
	if TRACE < l.level {
		return
	}

	l.write(TRACE, v...)
}

// IsLevel проверяет актуальность уровня логирования.
// Возвращает true, если указанный уровень логирования пишется в журнал.
func IsLevel(level Level) bool {
	return std.IsLevel(level)
}

// Error выводит сообщение об ошибке и завершает работу приложения.
// Пишет сообщение о фатальной ошибке и вызывает: os.Exit(1).
func Error(v ...interface{}) {
	std.Error(v...)
}

// Warn выводит предупреждение.
// Вызов игнорируется, если уровень важности логируемых сообщений не соответствует: WARN.
func Warn(v ...interface{}) {
	std.Warn(v...)
}

// Info выводит информационное сообщение.
// Вызов игнорируется, если уровень важности логируемых сообщений не соответствует: INFO.
func Info(v ...interface{}) {
	std.Info(v...)
}

// Debug выводит отладочное сообщение.
// Вызов игнорируется, если уровень важности логируемых сообщений не соответствует: DEBUG.
func Debug(v ...interface{}) {
	std.Debug(v...)
}

// Trace выводит произвольное сообщение.
// Вызов игнорируется, если уровень важности логируемых сообщений не соответствует: TRACE.
func Trace(v ...interface{}) {
	std.Trace(v...)
}

// IsError проверяет актуальность уровня логирования: ERROR.
// Возвращает true, если сообщения этого уровня пишутся в журнал.
func IsError() bool {
	return std.IsError()
}

// IsWarn проверяет актуальность уровня логирования: WARN.
// Возвращает true, если сообщения этого уровня пишутся в журнал.
func IsWarn() bool {
	return std.IsWarn()
}

// IsInfo проверяет актуальность уровня логирования: INFO.
// Возвращает true, если сообщения этого уровня пишутся в журнал.
func IsInfo() bool {
	return std.IsInfo()
}

// IsDebug проверяет актуальность уровня логирования: DEBUG.
// Возвращает true, если сообщения этого уровня пишутся в журнал.
func IsDebug() bool {
	return std.IsDebug()
}

// IsTrace проверяет актуальность уровня логирования: TRACE.
// Возвращает true, если сообщения этого уровня пишутся в журнал.
func IsTrace() bool {
	return std.IsTrace()
}
