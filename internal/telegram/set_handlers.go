package telegram

import (
	"cws/internal/telegram/messaging"
	"cws/logger"
)

func (bs *BotService) SetHandlers(msgSender messaging.MessageSender, cmdHdlr CommandHandler, callbackHdlr CallbackHandler, dialogHdlr DialogHandler, docHdlr DocumentHandler) {
	bs.msgSender = msgSender
	bs.cmdHdlr = cmdHdlr
	bs.callbackHdlr = callbackHdlr
	bs.dialogHdlr = dialogHdlr
	bs.docHdlr = docHdlr

	err := bs.setupMenuButton()
	if err != nil {
		logger.Warn("Не удалось установить Menu Button: %v", err)
	}
}
