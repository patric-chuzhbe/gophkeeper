package datatypehandler

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"os"
	"strings"
)

type BinaryHandler struct{}

func (h *BinaryHandler) Input(reader *bufio.Reader) ([]byte, error) {
	fmt.Print("Enter path to binary file: ")
	path, _ := reader.ReadString('\n')
	path = strings.TrimSpace(path)

	file, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return file, nil
}

func (h *BinaryHandler) Display(data []byte) {
	// base64 for safe printout
	encoded := base64.StdEncoding.EncodeToString(data)
	fmt.Println("Binary data (base64):")
	fmt.Println(encoded)
}
