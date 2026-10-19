package datatypehandler

import (
	"bufio"
	"encoding/json"
	"fmt"
	"strings"
)

type BankCard struct {
	Number   string `json:"number"`
	Expiry   string `json:"expiry"`
	CVV      string `json:"cvv"`
	FullName string `json:"name"`
}

type BankCardHandler struct{}

func (h *BankCardHandler) Input(reader *bufio.Reader) ([]byte, error) {
	fmt.Print("Card Number: ")
	number, _ := reader.ReadString('\n')

	fmt.Print("Expiry (MM/YY): ")
	expiry, _ := reader.ReadString('\n')

	fmt.Print("CVV: ")
	cvv, _ := reader.ReadString('\n')

	fmt.Print("Cardholder Name: ")
	name, _ := reader.ReadString('\n')

	obj := BankCard{
		Number:   strings.TrimSpace(number),
		Expiry:   strings.TrimSpace(expiry),
		CVV:      strings.TrimSpace(cvv),
		FullName: strings.TrimSpace(name),
	}
	return json.Marshal(obj)
}

func (h *BankCardHandler) Display(data []byte) {
	var obj BankCard
	if err := json.Unmarshal(data, &obj); err != nil {
		fmt.Println("Invalid card format")
		return
	}
	fmt.Printf("Card: %s  Expiry: %s\n", obj.Number, obj.Expiry)
	fmt.Printf("Name: %s  CVV: %s\n", obj.FullName, obj.CVV)
}
