package models

type UserRecord struct {
	Type  string
	Data  []byte
	Nonce []byte
}
