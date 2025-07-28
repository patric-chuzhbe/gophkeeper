package models

type UserRecord struct {
	Type  string
	Data  []byte
	Nonce []byte
	Meta  map[string]string
}

const (
	DataTypeLogPass  = "LOG_PASS"
	DataTypeText     = "TEXT"
	DataTypeBinary   = "BINARY"
	DataTypeBankCard = "BANK_CARD"
)
