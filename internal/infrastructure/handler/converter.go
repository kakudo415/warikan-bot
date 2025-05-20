package handler

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/kakudo415/warikan-bot/internal/domain/valueobject"
	"github.com/slack-go/slack"
)

func parseYen(text string) (valueobject.Yen, error) {
	rawYen := strings.ReplaceAll(text, ",", "")
	amount, err := strconv.Atoi(rawYen)
	if err != nil {
		return valueobject.Yen(0), fmt.Errorf("failed to parse amount: %w", err)
	}
	yen, err := valueobject.NewYen(amount)
	if err != nil {
		return valueobject.Yen(0), err
	}
	return yen, nil
}

func buildPaymentCreatedMessage(userID string, amount valueobject.Yen) slack.MsgOption {
	return slack.MsgOptionBlocks(
		slack.NewSectionBlock(
			slack.NewTextBlockObject("mrkdwn", fmt.Sprintf("<@%s>さんが%s立て替えました！", userID, amount.String()), false, false),
			nil,
			nil,
		),
	)
}

func buildPayerJoinedMessage(userID string) slack.MsgOption {
	return slack.MsgOptionBlocks(
		slack.NewSectionBlock(
			slack.NewTextBlockObject("mrkdwn", fmt.Sprintf("<@%s>さんが割り勘に参加します！", userID), false, false),
			nil,
			nil,
		),
	)
}

func buildPayerAlreadyJoinedMessage(userID string) slack.MsgOption {
	return slack.MsgOptionBlocks(
		slack.NewSectionBlock(
			slack.NewTextBlockObject("mrkdwn", fmt.Sprintf("<@%s>さんはすでに割り勘に参加してくれています！", userID), false, false),
			nil,
			nil,
		),
	)
}

func buildHelpMessage() slack.MsgOption {
	return slack.MsgOptionBlocks(
		slack.NewSectionBlock(
			slack.NewTextBlockObject("mrkdwn", "*Slackで割り勘の計算ができます* :tada:\n支払いの集計はチャンネルごとに行われるので、イベント用のチャンネルで使ってください！", false, false),
			nil,
			nil,
		),
		slack.NewDividerBlock(),
		slack.NewSectionBlock(
			slack.NewTextBlockObject("mrkdwn", ":moneybag: *立替え登録*", false, false),
			[]*slack.TextBlockObject{
				slack.NewTextBlockObject("mrkdwn", "*登録する*\n`/warikan [金額]`", false, false),
				slack.NewTextBlockObject("mrkdwn", "*取り消す*\n登録メッセージを削除してください", false, false),
			},
			nil,
		),
		slack.NewDividerBlock(),
		slack.NewSectionBlock(
			slack.NewTextBlockObject("mrkdwn", ":money_with_wings: *支払者登録*", false, false),
			[]*slack.TextBlockObject{
				slack.NewTextBlockObject("mrkdwn", "*登録する*\n`/warikan join`", false, false),
				slack.NewTextBlockObject("mrkdwn", "*取り消す*\n登録メッセージを削除してください", false, false),
			},
			nil,
		),
		slack.NewDividerBlock(),
		slack.NewSectionBlock(
			slack.NewTextBlockObject("mrkdwn", ":beginner: *ヘルプ*", false, false),
			[]*slack.TextBlockObject{
				slack.NewTextBlockObject("mrkdwn", "*この使い方を表示する*\n`/warikan help`", false, false),
			},
			nil,
		),
	)
}

func buildInvalidCommandMessage() slack.MsgOption {
	return slack.MsgOptionBlocks(
		slack.NewSectionBlock(
			slack.NewTextBlockObject("mrkdwn", "無効なコマンドです。使い方は `/warikan help` をご覧ください！", false, false),
			nil,
			nil,
		),
	)
}
