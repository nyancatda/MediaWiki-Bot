package Command

import (
	"github.com/nyancatda/ShionBot/Utils/Language"
)

func Help(SNSName string, UserID string) (string, bool) {
	HelpText := Language.Message(SNSName, UserID).HelpText
	return HelpText, true
}
