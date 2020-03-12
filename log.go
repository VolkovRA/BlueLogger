// Package log - Цветной логгер.
// По сути, это обычный golang логгер, только с добавлением уровней логгирования и возможностью цветного оформления текста.
package log

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/VolkovRA/acolor"
)

// LogLevel - Уровень важности логируемых сообщений.
// Необходим для разделения сообщений журнала по уровню важности.
// Все доступные значения и их описания для применения перечислены в константах пакета.
type LogLevel int32

// Уровни важности логируемых сообщений.
// Тут перечислены все доступные уровни важности и их описание для применения.
const (

	// LevelError - Критическая ошибка.
	// Приложение в критическом состоянии, требуется вмешательство человека.
	LevelError LogLevel = iota

	// LevelWarn - Предупреждение.
	// Возникли осложнения, но программа умная и смогла их разрешить самостоятельно.
	LevelWarn

	// LevelInfo - Общая информация.
	// Что сейчас происходит.
	LevelInfo

	// LevelDebug - Отладочные сообщения.
	// Дополнительная информация, которая может быть полезна для диагностики или отладки.
	LevelDebug

	// LevelTrace - Всё подряд.
	// Пишем всё подряд.
	LevelTrace
)

// Дефолтный логгер
var std = New(os.Stderr, LevelTrace)

// Logger логгер сообщений.
// Используется для вывода сообщений в журнал приложения.
// Вы можете создать новый экземпляр логгера или использовать дефолтный, на уровне пакета.
type Logger struct {

	// Цветной текст.
	// Если true, логгер подкрашивает каждое сообщения.
	// Работает на основе добавления управляющих ANSI символов, не работает в Windows.
	// По умолчанию: true.
	Color bool

	// Время в UTC.
	// Если true, логгер будет использовать нулевой часовой пояс, установленный в локальной системе.
	// По умолчанию: true.
	UTC bool

	// Отображение заголовка. (Целиком)
	// Если true, логгер добавляет в каждое сообщение заголовок с системной информацией: время, уровень важности и т.п.
	// По умолчанию: true.
	Head bool

	// Отображение уровня важности в заголовке.
	// Если true, в заголовке каждого сообщения будет присутствовать маркер уровня важности данного сообщения: [LEVEL].
	// По умолчанию: true.
	HeadLevel bool

	// Отображение даты в заголовке.
	// Если true, в заголовке каждого сообщения будет присутствовать дата: DD.MM.YYYY.
	// По умолчанию: true.
	HeadDate bool

	// Отображение времени в заголовке.
	// Если true, в заголовке каждого сообщения будет присутствовать время: HH:MM:SS.
	// По умолчанию: true.
	HeadTime bool

	// Отображение микросекунд в заголовке. (Работает только при включенном HeadTime)
	// Если true, в заголовке каждого сообщения будут присутствовать микросекунды: HH:MM:SS.000000
	// По умолчанию: false.
	HeadMC bool

	mu    sync.Mutex // Атомарная запись.
	out   io.Writer  // Назначение для вывода сообщений.
	level LogLevel   // Уровень логируемых сообщений.
	buf   []byte     // Буфер для сложения текста при записи.
}

// New создаёт новый логгер.
// Вы можете указать цель назначения всех сообщений журнала.
func New(out io.Writer, level LogLevel) *Logger {
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
func (l *Logger) writeHeader(buf *[]byte, level LogLevel) {

	// Метка уровня:
	if l.HeadLevel {
		*buf = append(*buf, l.getHeaderLevel(level)...)
	}

	// Цвет заголовка:
	if l.Color {
		if level == LevelError {
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
func (l *Logger) getHeaderLevel(level LogLevel) string {
	if l.Color {
		switch level {
		case LevelInfo:
			return acolor.Apply(acolor.Bold, acolor.Green) + "[INFO]  " + acolor.Clear()
		case LevelWarn:
			return acolor.Apply(acolor.Bold, acolor.Yellow) + "[WARN]  " + acolor.Clear()
		case LevelTrace:
			return acolor.Apply(acolor.Bold, acolor.White) + "[TRACE] " + acolor.Clear()
		case LevelDebug:
			return acolor.Apply(acolor.Bold, acolor.Cyan) + "[DEBUG] " + acolor.Clear()
		case LevelError:
			return acolor.Apply(acolor.Bold, acolor.Red) + "[ERROR] " + acolor.Clear()
		default:
			return acolor.Apply(acolor.Bold, acolor.White) + "[]      " + acolor.Clear()
		}
	} else {
		switch level {
		case LevelInfo:
			return "[INFO]  "
		case LevelWarn:
			return "[WARN]  "
		case LevelTrace:
			return "[TRACE] "
		case LevelDebug:
			return "[DEBUG] "
		case LevelError:
			return "[ERROR] "
		default:
			return "[]      "
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
func (l *Logger) write(level LogLevel, v ...interface{}) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Шапка:
	l.buf = l.buf[:0]
	if l.Head {
		l.writeHeader(&l.buf, level)
	}

	// Тело:
	if l.Color && level == LevelError {
		l.buf = append(l.buf, (acolor.Apply(acolor.Red) + fmt.Sprint(v...) + acolor.Clear() + "\n")...)
	} else {
		l.buf = append(l.buf, (fmt.Sprint(v...) + "\n")...)
	}

	// Вывод:
	_, err := l.out.Write(l.buf)

	return err
}

// Записать сообщение в журнал с применением форматирования.
func (l *Logger) writef(level LogLevel, format string, v ...interface{}) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Шапка:
	l.buf = l.buf[:0]
	if l.Head {
		l.writeHeader(&l.buf, level)
	}

	// Тело:
	if l.Color && level == LevelError {
		l.buf = append(l.buf, (acolor.Apply(acolor.Red) + fmt.Sprintf(format, v...) + acolor.Clear() + "\n")...)
	} else {
		l.buf = append(l.buf, (fmt.Sprintf(format, v...) + "\n")...)
	}

	// Вывод:
	_, err := l.out.Write(l.buf)

	return err
}

// Level уровень важности логируемых сообщений.
// Если сообщение не соответствует уровню важности, оно не попадает в журнал.
// По умолчанию: LevelTrace. (В журнал попадают все сообщения)
func (l *Logger) Level() LogLevel {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.level
}

// SetLevel устанавливает уровень важности логируемых сообщений.
// Доступные значения LogLevel смотрите в константах пакета.
func (l *Logger) SetLevel(level LogLevel) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if level < LevelError {
		l.level = LevelError
		return
	}
	if level > LevelTrace {
		l.level = LevelTrace
		return
	}

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

// IsLevel проверяет актуальность уровня логирования.
// Возвращает true, если указанный уровень логирования пишется в журнал.
func (l *Logger) IsLevel(level LogLevel) bool {
	return l.level >= level
}

// IsError проверяет актуальность уровня логгирования: ERROR.
// Возвращает true, если сообщения этого уровня пишутся в журнал.
func (l *Logger) IsError() bool {
	return l.IsLevel(LevelError)
}

// IsWarn проверяет актуальность уровня логгирования: WARN.
// Возвращает true, если сообщения этого уровня пишутся в журнал.
func (l *Logger) IsWarn() bool {
	return l.IsLevel(LevelWarn)
}

// IsInfo проверяет актуальность уровня логгирования: INFO.
// Возвращает true, если сообщения этого уровня пишутся в журнал.
func (l *Logger) IsInfo() bool {
	return l.IsLevel(LevelInfo)
}

// IsDebug проверяет актуальность уровня логгирования: DEBUG.
// Возвращает true, если сообщения этого уровня пишутся в журнал.
func (l *Logger) IsDebug() bool {
	return l.IsLevel(LevelDebug)
}

// IsTrace проверяет актуальность уровня логгирования: TRACE.
// Возвращает true, если сообщения этого уровня пишутся в журнал.
func (l *Logger) IsTrace() bool {
	return l.IsLevel(LevelTrace)
}

// Error выводит сообщение об ошибке и завершает работу приложения.
// Пишет сообщение о фатальной ошибке и вызывает: os.Exit(1).
func (l *Logger) Error(v ...interface{}) {
	l.write(LevelError, v...)
	os.Exit(1)
}

// Warn выводит предупреждение.
// Вызов игнорируется, если уровень важности логируемых сообщений не соответствует: WARN.
func (l *Logger) Warn(v ...interface{}) {
	if l.level < LevelWarn {
		return
	}

	l.write(LevelWarn, v...)
}

// Info выводит информационное сообщение.
// Вызов игнорируется, если уровень важности логируемых сообщений не соответствует: INFO.
func (l *Logger) Info(v ...interface{}) {
	if l.level < LevelInfo {
		return
	}

	l.write(LevelInfo, v...)
}

// Debug выводит отладочное сообщение.
// Вызов игнорируется, если уровень важности логируемых сообщений не соответствует: DEBUG.
func (l *Logger) Debug(v ...interface{}) {
	if l.level < LevelDebug {
		return
	}

	l.write(LevelDebug, v...)
}

// Trace выводит произвольное сообщение.
// Вызов игнорируется, если уровень важности логируемых сообщений не соответствует: TRACE.
func (l *Logger) Trace(v ...interface{}) {
	if l.level < LevelTrace {
		return
	}

	l.write(LevelTrace, v...)
}

// Errorf выводит сообщение об ошибке с применением форматирования и завершает работу приложения.
// Пишет сообщение о фатальной ошибке и вызывает: os.Exit(1).
func (l *Logger) Errorf(format string, v ...interface{}) {
	l.writef(LevelError, format, v...)
	os.Exit(1)
}

// Warnf выводит предупреждение с применением форматирования.
// Вызов игнорируется, если уровень важности логируемых сообщений не соответствует: WARN.
func (l *Logger) Warnf(format string, v ...interface{}) {
	if l.level < LevelWarn {
		return
	}

	l.writef(LevelWarn, format, v...)
}

// Infof выводит информационное сообщение с применением форматирования.
// Вызов игнорируется, если уровень важности логируемых сообщений не соответствует: INFO.
func (l *Logger) Infof(format string, v ...interface{}) {
	if l.level < LevelInfo {
		return
	}

	l.writef(LevelInfo, format, v...)
}

// Debugf выводит отладочное сообщение с применением форматирования.
// Вызов игнорируется, если уровень важности логируемых сообщений не соответствует: DEBUG.
func (l *Logger) Debugf(format string, v ...interface{}) {
	if l.level < LevelDebug {
		return
	}

	l.writef(LevelDebug, format, v...)
}

// Tracef выводит произвольное сообщение с применением форматирования.
// Вызов игнорируется, если уровень важности логируемых сообщений не соответствует: TRACE.
func (l *Logger) Tracef(format string, v ...interface{}) {
	if l.level < LevelTrace {
		return
	}

	l.writef(LevelTrace, format, v...)
}

// IsLevel проверяет актуальность уровня логирования.
// Возвращает true, если указанный уровень логирования пишется в журнал.
func IsLevel(level LogLevel) bool {
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

// Errorf выводит сообщение об ошибке и завершает работу приложения с применением форматирования.
// Пишет сообщение о фатальной ошибке и вызывает: os.Exit(1).
func Errorf(format string, v ...interface{}) {
	std.Errorf(format, v...)
}

// Warnf выводит предупреждение с применением форматирования.
// Вызов игнорируется, если уровень важности логируемых сообщений не соответствует: WARN.
func Warnf(format string, v ...interface{}) {
	std.Warnf(format, v...)
}

// Infof выводит информационное сообщение с применением форматирования.
// Вызов игнорируется, если уровень важности логируемых сообщений не соответствует: INFO.
func Infof(format string, v ...interface{}) {
	std.Infof(format, v...)
}

// Debugf выводит отладочное сообщение с применением форматирования.
// Вызов игнорируется, если уровень важности логируемых сообщений не соответствует: DEBUG.
func Debugf(format string, v ...interface{}) {
	std.Debugf(format, v...)
}

// Tracef выводит произвольное сообщение с применением форматирования.
// Вызов игнорируется, если уровень важности логируемых сообщений не соответствует: TRACE.
func Tracef(format string, v ...interface{}) {
	std.Tracef(format, v...)
}

// IsError проверяет актуальность уровня логгирования: ERROR.
// Возвращает true, если сообщения этого уровня пишутся в журнал.
func IsError() bool {
	return std.IsError()
}

// IsWarn проверяет актуальность уровня логгирования: WARN.
// Возвращает true, если сообщения этого уровня пишутся в журнал.
func IsWarn() bool {
	return std.IsWarn()
}

// IsInfo проверяет актуальность уровня логгирования: INFO.
// Возвращает true, если сообщения этого уровня пишутся в журнал.
func IsInfo() bool {
	return std.IsInfo()
}

// IsDebug проверяет актуальность уровня логгирования: DEBUG.
// Возвращает true, если сообщения этого уровня пишутся в журнал.
func IsDebug() bool {
	return std.IsDebug()
}

// IsTrace проверяет актуальность уровня логгирования: TRACE.
// Возвращает true, если сообщения этого уровня пишутся в журнал.
func IsTrace() bool {
	return std.IsTrace()
}
