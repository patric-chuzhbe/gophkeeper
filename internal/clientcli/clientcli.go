package clientcli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/patric-chuzhbe/gophkeeper/internal/version"

	"github.com/patric-chuzhbe/gophkeeper/internal/clientvault"
	"github.com/spf13/cobra"
)

type CLI struct {
	rootCmd *cobra.Command
	vault   vault
}

type vault interface {
	AddRecord(password string, record clientvault.VaultRecord) error
	FetchRecords(password string) ([]clientvault.VaultRecord, error)
}

func New(vault vault) *CLI {
	cli := &CLI{
		vault: vault,
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

	return cli
}

func (c *CLI) Execute() {
	cobra.CheckErr(c.rootCmd.Execute())
}

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

func (c *CLI) Get(cmd *cobra.Command, args []string) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Master password: ")
	password, _ := reader.ReadString('\n')
	password = password[:len(password)-1]

	records, err := c.vault.FetchRecords(password)
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
		fmt.Printf("  Decrypted Data: %s\n", string(r.Data))
	}
}

func (c *CLI) Version(cmd *cobra.Command, args []string) {
	fmt.Println(version.Info())
}
