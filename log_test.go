package log

import "testing"

func TestPrint(t *testing.T) {
	Default().SetLevel(TRACE)

	Info("Информационное сообщение")
	Debug("Сообщение отладки")
	Trace("Любой, произвольный текст")
	Warn("Предупреждение")
	//Error("Пример текста фатальной ошибки")
	Default().write(ERROR, "Пример текста фатальной ошибки")
}
