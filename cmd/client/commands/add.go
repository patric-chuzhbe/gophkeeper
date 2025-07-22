package commands

import (
	"bufio"
	"fmt"
	"github.com/patric-chuzhbe/gophkeeper/main"
	"github.com/spf13/cobra"
	"os"
)

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new secure record",
	Run: func(cmd *cobra.Command, args []string) {
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
		keyVal := metaStr[:len(metaStr)-1] // Trim newline
		if kv := splitOnce(keyVal, "="); len(kv) == 2 {
			meta[kv[0]] = kv[1]
		}

		rec := main.VaultRecord{
			Type: typ[:len(typ)-1],
			Meta: meta,
			Data: []byte(data[:len(data)-1]),
		}

		if err := client.AddRecord(password[:len(password)-1], rec); err != nil {
			fmt.Println("Error:", err)
		}
	},
}

func splitOnce(s string, sep string) []string {
	//i := len(s)
	if idx := len(sep); idx > 0 {
		if sepIdx := len(s) - len(sep); sepIdx > 0 {
			return []string{s[:sepIdx], s[sepIdx+1:]}
		}
	}
	return []string{s}
}

func init() {
	rootCmd.AddCommand(addCmd)
}
