package acct

import (
	"fmt"
	"strings"

	"github.com/blebbit/at-mirror/pkg/sql"
	"github.com/spf13/cobra"
)

func getValidSQLs(prefix string) ([]string, error) {
	files, err := sql.SQLFiles.ReadDir(prefix)
	if err != nil {
		return nil, err
	}
	var validArgs []string
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".sql") {
			arg := strings.TrimSuffix(file.Name(), ".sql")
			validArgs = append(validArgs, arg)
		}
	}
	return validArgs, nil
}

func getValidArgs(prefix string) func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		validSQLs, err := getValidSQLs(prefix)
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}

		var completions []string
		for _, arg := range validSQLs {
			if strings.HasPrefix(arg, toComplete) {
				completions = append(completions, arg)
			}
		}
		return completions, cobra.ShellCompDirectiveNoFileComp
	}
}

func listValidSQLs(prefix string) {
	sqls, err := getValidSQLs(prefix)
	if err != nil {
		fmt.Println("Error getting sqls:", err)
		return
	}
	for _, sql := range sqls {
		fmt.Println(sql)
	}
}

func viewSQL(prefix, sqlName string) {
	path := fmt.Sprintf("%s/%s.sql", prefix, sqlName)
	content, err := sql.SQLFiles.ReadFile(path)
	if err != nil {
		fmt.Printf("Error reading sql %s: %v\n", sqlName, err)
		return
	}
	fmt.Println(string(content))
}
