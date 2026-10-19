package datatypehandler

import (
	"bufio"
	"fmt"
	"strings"
)

type TextHandler struct{}

func (h *TextHandler) Input(reader *bufio.Reader) ([]byte, error) {
	fmt.Print("Enter text: ")
	text, _ := reader.ReadString('\n')
	return []byte(strings.TrimSpace(text)), nil
}

func (h *TextHandler) Display(data []byte) {
	fmt.Println("Text:")
	fmt.Println(string(data))
}
