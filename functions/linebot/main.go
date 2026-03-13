// Package linebot は LINE Messaging API の Webhook を受け取り、
// PDF にアレルゲンハイライトを施して画像で返す Cloud Function を提供します。
package linebot

import (
	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
)

func init() {
	functions.HTTP("LineBotWebhook", Handler)
}
