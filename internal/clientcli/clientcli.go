package clientcli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	//"bufio"
	//"fmt"
	//"github.com/patric-chuzhbe/gophkeeper/internal/config"
	"github.com/patric-chuzhbe/gophkeeper/internal/clientvault"
	"github.com/spf13/cobra"
	//"os"
)

type CLI struct {
	rootCmd       *cobra.Command
	vault         vault
	cryptographer cryptographer
}

type vault interface {
	AddRecord(password string, record clientvault.VaultRecord) error
}

type cryptographer interface {
	DeriveKey(password, salt []byte) []byte
	Encrypt(plaintext []byte, key []byte) ([]byte, []byte, error)
}

/*< new methods>*/
func (c *CLI) Add(cmd *cobra.Command, args []string) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Master password: ")
	password, _ := reader.ReadString('\n')

	fmt.Print("Type (e.g. login_password): ")
	typ, _ := reader.ReadString('\n')

	fmt.Print("Meta (e.g. website=example.com): ")
	metaStr, _ := reader.ReadString('\n')

	fmt.Print("Data to encrypt: ")
	data, _ := reader.ReadString('\n')

	meta := map[string]string{}
	keyVal := metaStr[:len(metaStr)-1]
	if kv := strings.SplitN(keyVal, "=", 2); len(kv) == 2 {
		meta[kv[0]] = kv[1]
	}

	rec := clientvault.VaultRecord{
		Type: typ[:len(typ)-1],
		Meta: meta,
		Data: []byte(data[:len(data)-1]),
	}

	if err := c.vault.AddRecord(password[:len(password)-1], rec); err != nil {
		fmt.Println("Error:", err)
	}
}

func (c *CLI) Execute() {
	cobra.CheckErr(c.rootCmd.Execute())
}

func New(vault vault, cryptographer cryptographer) *CLI {
	cli := &CLI{
		vault:         vault,
		cryptographer: cryptographer,
		rootCmd: &cobra.Command{
			Use:   "vault-cli",
			Short: "Secure CLI client for managing encrypted data",
		},
	}

	cli.rootCmd.AddCommand(&cobra.Command{
		Use:   "add",
		Short: "Add a new secure record",
		Run:   cli.Add,
	})

	return cli
}

/*</new methods>*/
