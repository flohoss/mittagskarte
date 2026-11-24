package telegram

import (
	"ganigeorgiev.com/pocketbase/pocketbase/core"
)

const activeField = "active"

func BindHooks(app core.App) {
	app.OnRecordCreate("telegram_chats").BindFunc(func(e *core.RecordEvent) error {
		e.Record.Set(activeField, true)
		return e.Next()
	})
}
