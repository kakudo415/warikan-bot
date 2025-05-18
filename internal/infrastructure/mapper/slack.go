package mapper

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"

	"github.com/kakudo415/warikan-bot/internal/domain/valueobject"
)

type SlackMapper struct {
	db *sql.DB
}

func NewSlackMapper(filename string) (*SlackMapper, error) {
	db, err := sql.Open("sqlite3", filename)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS events_slack_channels (
			event_id TEXT PRIMARY KEY,
			slack_channel_id TEXT NOT NULL
		);
		CREATE TABLE IF NOT EXISTS payers_slack_users (
			payer_id TEXT PRIMARY KEY,
			slack_user_id TEXT NOT NULL
		);
		CREATE TABLE IF NOT EXISTS payments_slack_threads (
			payment_id TEXT PRIMARY KEY,
			slack_thread_ts TEXT NOT NULL
		);
	`)
	if err != nil {
		return nil, err
	}

	return &SlackMapper{
		db: db,
	}, nil
}

func (m *SlackMapper) EventIDToSlackChannelID(eventID valueobject.EventID) (string, error) {
	var slackChannelID string
	err := m.db.QueryRow("SELECT slack_channel_id FROM events_slack_channels WHERE event_id = ?", eventID.String()).Scan(&slackChannelID)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", err
	}
	return slackChannelID, nil
}

func (m *SlackMapper) SlackChannelIDToEventID(slackChannelID string) (valueobject.EventID, error) {
	var rawEventID string
	err := m.db.QueryRow("SELECT event_id FROM events_slack_channels WHERE slack_channel_id = ?", slackChannelID).Scan(&rawEventID)
	if err != nil {
		if err == sql.ErrNoRows {
			return valueobject.EventIDUnknown, nil
		}
		return valueobject.EventIDUnknown, err
	}
	eventID, err := valueobject.EventIDFromString(rawEventID)
	if err != nil {
		return valueobject.EventIDUnknown, err
	}
	return valueobject.EventID(eventID), nil
}

func (m *SlackMapper) CreateEventIDSlackChannelIDMapping(eventID valueobject.EventID, slackChannelID string) error {
	_, err := m.db.Exec("INSERT INTO events_slack_channels (event_id, slack_channel_id) VALUES (?, ?)",
		eventID.String(),
		slackChannelID,
	)
	return err
}

func (m *SlackMapper) PayerIDToSlackUserID(payerID valueobject.PayerID) (string, error) {
	var slackUserID string
	err := m.db.QueryRow("SELECT slack_user_id FROM payers_slack_users WHERE payer_id = ?", payerID.String()).Scan(&slackUserID)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", err
	}
	return slackUserID, nil
}

func (m *SlackMapper) SlackUserIDToPayerID(slackUserID string) (valueobject.PayerID, error) {
	var rawPayerID string
	err := m.db.QueryRow("SELECT payer_id FROM payers_slack_users WHERE slack_user_id = ?", slackUserID).Scan(&rawPayerID)
	if err != nil {
		if err == sql.ErrNoRows {
			return valueobject.PayerIDUnknown, nil
		}
		return valueobject.PayerIDUnknown, err
	}
	payerID, err := valueobject.PayerIDFromString(rawPayerID)
	if err != nil {
		return valueobject.PayerIDUnknown, err
	}
	return valueobject.PayerID(payerID), nil
}

func (m *SlackMapper) CreatePayerIDSlackUserIDMapping(payerID valueobject.PayerID, slackUserID string) error {
	_, err := m.db.Exec("INSERT INTO payers_slack_users (payer_id, slack_user_id) VALUES (?, ?)",
		payerID.String(),
		slackUserID,
	)
	return err
}

func (m *SlackMapper) PaymentIDToSlackThreadTS(paymentID valueobject.PaymentID) (string, error) {
	var slackThreadTS string
	err := m.db.QueryRow("SELECT slack_thread_ts FROM payments_slack_threads WHERE payment_id = ?", paymentID.String()).Scan(&slackThreadTS)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", err
	}
	return slackThreadTS, nil
}

func (m *SlackMapper) SlackThreadTSToPaymentID(slackThreadTS string) (valueobject.PaymentID, error) {
	var rawPaymentID string
	err := m.db.QueryRow("SELECT payment_id FROM payments_slack_threads WHERE slack_thread_ts = ?", slackThreadTS).Scan(&rawPaymentID)
	if err != nil {
		if err == sql.ErrNoRows {
			return valueobject.PaymentIDUnknown, nil
		}
		return valueobject.PaymentIDUnknown, err
	}
	paymentID, err := valueobject.PaymentIDFromString(rawPaymentID)
	if err != nil {
		return valueobject.PaymentIDUnknown, err
	}
	return valueobject.PaymentID(paymentID), nil
}

func (m *SlackMapper) CreatePaymentIDSlackThreadTSMapping(paymentID valueobject.PaymentID, slackThreadTS string) error {
	_, err := m.db.Exec("INSERT INTO payments_slack_threads (payment_id, slack_thread_ts) VALUES (?, ?)",
		paymentID.String(),
		slackThreadTS,
	)
	return err
}
