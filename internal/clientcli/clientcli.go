package clientcli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/patric-chuzhbe/gophkeeper/internal/datatypehandler"
	"github.com/patric-chuzhbe/gophkeeper/internal/models"

	"github.com/spf13/cobra"

	"github.com/patric-chuzhbe/gophkeeper/internal/version"

	"github.com/patric-chuzhbe/gophkeeper/internal/clientvault"
)

type CLI struct {
	rootCmd *cobra.Command
	vault   vault
}

type vault interface {
	AddRecord(record clientvault.VaultRecord) error
	FetchRecords() ([]clientvault.VaultRecord, error)
	Login(login string, password string) error
	Register(login, password string) error
}

/*< new funcs>*/

type dataTypeHandler interface {
	Input(reader *bufio.Reader) ([]byte, error)
	Display(data []byte /*, meta map[string]string*/)
}

func getHandler(typ string) dataTypeHandler {
	switch typ {
	case models.DataTypeLogPass /*"LOG_PASS"*/ :
		return &datatypehandler.LoginPassHandler{}
	case models.DataTypeText /*"TEXT"*/ :
		return &datatypehandler.TextHandler{}
	case models.DataTypeBinary /*"BINARY"*/ :
		return &datatypehandler.BinaryHandler{}
	case models.DataTypeBankCard /*"BANK_CARD"*/ :
		return &datatypehandler.BankCardHandler{}
	default:
		return &datatypehandler.TextHandler{} // fallback
	}
}

func (c *CLI) Register(cmd *cobra.Command, args []string) {
	reader := bufio.NewReader(os.Stdin)

	login := readLine(reader, "Login: ")
	password := readLine(reader, "Password: ")

	if err := c.vault.Register(login, password); err != nil {
		fmt.Println("Error:", err)
	}

	if err := c.vault.Login(login, password); err != nil {
		fmt.Println("Login failed:", err)
	}
}

func (c *CLI) Login(cmd *cobra.Command, args []string) {
	reader := bufio.NewReader(os.Stdin)
	login := readLine(reader, "Login: ")
	password := readLine(reader, "Password: ")

	if err := c.vault.Login(login, password); err != nil {
		fmt.Println("Login failed:", err)
	}
}

func New(vault vault) *CLI {
	cli := &CLI{
		vault: vault,
		rootCmd: &cobra.Command{
			Use:   "client",
			Short: "Secure CLI client for managing encrypted data",
		},
	}

	cli.rootCmd.AddCommand(&cobra.Command{
		Use:   "add",
		Short: "Add a new secure record",
		Run:   cli.Add,
	})

	cli.rootCmd.AddCommand(&cobra.Command{
		Use:   "get",
		Short: "Fetch and decrypt records from server",
		Run:   cli.Get,
	})

	cli.rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print version and build date",
		Run:   cli.Version,
	})

	cli.rootCmd.AddCommand(&cobra.Command{
		Use:   "login",
		Short: "Authenticate and get token",
		Run:   cli.Login,
	})

	cli.rootCmd.AddCommand(&cobra.Command{
		Use:   "register",
		Short: "Register a new user",
		Run:   cli.Register,
	})

	return cli
}

func (c *CLI) Execute() {
	cobra.CheckErr(c.rootCmd.Execute())
}

func (c *CLI) Add(cmd *cobra.Command, args []string) {
	reader := bufio.NewReader(os.Stdin)

	typ := readLine(reader, "Type (e.g. login_password): ")
	metaStr := readLine(reader, "Meta (e.g. website=example.com; person = John Smith): ")
	//data := readLine(reader, "Data to encrypt: ")

	meta := map[string]string{}
	pairs := strings.Split(metaStr, ";")
	for _, pair := range pairs {
		if kv := strings.SplitN(pair, "=", 2); len(kv) == 2 {
			meta[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
		}
	}

	handler := getHandler(typ)
	data, err := handler.Input(reader)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	rec := clientvault.VaultRecord{
		Type:                  typ,
		Meta:                  meta,
		Data:/*[]byte(*/ data, /*)*/
	}

	if err := c.vault.AddRecord(rec); err != nil {
		fmt.Println("Error:", err)
	}
}

func (c *CLI) Get(cmd *cobra.Command, args []string) {
	records, err := c.vault.FetchRecords()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	if len(records) == 0 {
		fmt.Println("No accessible records.")
		return
	}

	for i, r := range records {
		fmt.Printf("\n[%d] Type: %s\n", i+1, r.Type)
		for k, v := range r.Meta {
			fmt.Printf("  %s: %s\n", k, v)
		}
		//fmt.Printf("  Decrypted Data: %s\n", string(r.Data))
		fmt.Println("  Decrypted Data:")
		handler := getHandler(r.Type)
		handler.Display(r.Data /*, r.Meta*/)
	}
}

func (c *CLI) Version(cmd *cobra.Command, args []string) {
	fmt.Println(version.Info())
}

func readLine(r *bufio.Reader, prompt string) string {
	fmt.Print(prompt)
	line, _ := r.ReadString('\n')
	return strings.TrimSpace(line)
}

/*</new funcs>*/
