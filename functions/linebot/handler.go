package linebot

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/line/line-bot-sdk-go/v8/linebot"
)

// fixedUserID は Firestore からアレルゲン情報を取得する際に使用する固定ユーザー ID です。
const fixedUserID = "default"

var bot *linebot.Client

func init() {
	var err error
	bot, err = linebot.New(
		os.Getenv("LINE_CHANNEL_SECRET"),
		os.Getenv("LINE_CHANNEL_ACCESS_TOKEN"),
	)
	if err != nil {
		panic(fmt.Sprintf("linebot クライアント初期化失敗: %v", err))
	}
}

// Handler は Cloud Functions の HTTP エントリーポイントです。
func Handler(w http.ResponseWriter, r *http.Request) {
	events, err := bot.ParseRequest(r)
	if err != nil {
		if err == linebot.ErrInvalidSignature {
			http.Error(w, "invalid signature", http.StatusBadRequest)
			return
		}
		http.Error(w, "parse request error", http.StatusInternalServerError)
		return
	}

	for _, event := range events {
		if err := handleEvent(r.Context(), event); err != nil {
			slog.Error("イベント処理エラー", "err", err)
		}
	}
	w.WriteHeader(http.StatusOK)
}

func handleEvent(ctx context.Context, event *linebot.Event) error {
	if event.Type != linebot.EventTypeMessage {
		return nil
	}

	// テキストメッセージはアレルゲン登録として処理する
	if textMsg, ok := event.Message.(*linebot.TextMessage); ok {
		return handleAllergenRegistration(ctx, event.ReplyToken, textMsg.Text)
	}

	fileMsg, ok := event.Message.(*linebot.FileMessage)
	if !ok {
		return nil
	}

	// PDF ファイルのみ処理する
	if !strings.HasSuffix(strings.ToLower(fileMsg.FileName), ".pdf") {
		_, err := bot.ReplyMessage(event.ReplyToken,
			linebot.NewTextMessage("PDF ファイルを送信してください。"),
		).Do()
		return err
	}

	slog.Info("ファイル受信", "file_name", fileMsg.FileName, "file_size", fileMsg.FileSize)

	// LINE Content API からファイルをダウンロードする
	content, err := bot.GetMessageContent(fileMsg.ID).Do()
	if err != nil {
		return fmt.Errorf("コンテンツ取得: %w", err)
	}
	defer content.Content.Close()

	pdfBytes, err := io.ReadAll(content.Content)
	if err != nil {
		return fmt.Errorf("ファイル読み込み: %w", err)
	}

	// highlight-pdf を呼び出す
	highlighted, err := callHighlightPDF(ctx, pdfBytes, fixedUserID)
	if err != nil {
		slog.Error("highlight-pdf 呼び出しエラー", "err", err)
		_, replyErr := bot.ReplyMessage(event.ReplyToken,
			linebot.NewTextMessage("PDF の処理中にエラーが発生しました。"),
		).Do()
		if replyErr != nil {
			slog.Error("エラー返信失敗", "err", replyErr)
		}
		return fmt.Errorf("highlight-pdf 呼び出し: %w", err)
	}

	// PDF を画像に変換する
	images, err := pdfToJPEGs(highlighted)
	if err != nil {
		return fmt.Errorf("PDF → 画像変換: %w", err)
	}

	// 画像を GCS にアップロードして公開 URL を取得する
	urls, err := uploadImages(ctx, images)
	if err != nil {
		return fmt.Errorf("画像アップロード: %w", err)
	}

	// LINE に画像を返信する（最大 5 枚）
	msgs := buildImageMessages(urls)
	_, err = bot.ReplyMessage(event.ReplyToken, msgs...).Do()
	return err
}

// handleAllergenRegistration はテキストメッセージをアレルゲン情報として Firestore に登録します。
func handleAllergenRegistration(ctx context.Context, replyToken, text string) error {
	target := strings.TrimSpace(text)
	if target == "" {
		return nil
	}

	if err := callRegisterAllergen(ctx, target, fixedUserID); err != nil {
		slog.Error("アレルゲン登録エラー", "err", err)
		_, replyErr := bot.ReplyMessage(replyToken,
			linebot.NewTextMessage("アレルゲン情報の登録中にエラーが発生しました。"),
		).Do()
		if replyErr != nil {
			slog.Error("エラー返信失敗", "err", replyErr)
		}
		return fmt.Errorf("アレルゲン登録: %w", err)
	}

	// 登録内容を箇条書きで表示する
	lines := strings.Split(target, "\n")
	var items []string
	for _, l := range lines {
		if l = strings.TrimSpace(l); l != "" {
			items = append(items, "・"+l)
		}
	}
	replyText := "以下のアレルゲン情報を登録しました。\n" + strings.Join(items, "\n")

	_, err := bot.ReplyMessage(replyToken,
		linebot.NewTextMessage(replyText),
	).Do()
	return err
}

// buildImageMessages は画像 URL から LINE メッセージを作成します（最大 5 件）。
func buildImageMessages(urls []string) []linebot.SendingMessage {
	const maxMessages = 5
	if len(urls) > maxMessages {
		urls = urls[:maxMessages]
	}
	msgs := make([]linebot.SendingMessage, len(urls))
	for i, url := range urls {
		msgs[i] = linebot.NewImageMessage(url, url)
	}
	return msgs
}
