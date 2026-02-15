package command

import (
	"fmt"

	"github.com/t3yamoto/gt/internal/auth"
	"github.com/urfave/cli/v2"
)

func LogoutCommand() *cli.Command {
	return &cli.Command{
		Name:  "logout",
		Usage: "保存されたトークンを削除",
		Action: func(c *cli.Context) error {
			if err := auth.DeleteToken(); err != nil {
				return err
			}
			fmt.Println("ログアウトしました。")
			return nil
		},
	}
}
