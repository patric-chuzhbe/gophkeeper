package datatypehandler

import (
	"bufio"
	"encoding/json"
	"fmt"
	"strings"
)

type LoginPassHandler struct{}

type LoginPass struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func (h *LoginPassHandler) Input(reader *bufio.Reader) ([]byte, error) {
	fmt.Print("Login: ")
	login, _ := reader.ReadString('\n')

	fmt.Print("Password: ")
	password, _ := reader.ReadString('\n')

	obj := LoginPass{
		Login:    strings.TrimSpace(login),
		Password: strings.TrimSpace(password),
	}

	return json.Marshal(obj)
}

func (h *LoginPassHandler) Display(data []byte) {
	var obj LoginPass
	if err := json.Unmarshal(data, &obj); err != nil {
		fmt.Println("Invalid login/pass format")
		return
	}
	fmt.Println("Login:   ", obj.Login)
	fmt.Println("Password:", obj.Password)
}
