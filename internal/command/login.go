package command

import (
	"fmt"

	"github.com/t3yamoto/gt/internal/auth"
	"github.com/urfave/cli/v2"
)

func LoginCommand() *cli.Command {
	return &cli.Command{
		Name:  "login",
		Usage: "OAuth認証を実行しトークンを保存",
		Action: func(c *cli.Context) error {
			if auth.HasToken() {
				fmt.Println("既にログイン済みです。再認証するには `gt logout` を実行してからやり直してください。")
				return nil
			}

			token, err := auth.Authenticate(c.Context)
			if err != nil {
				return fmt.Errorf("認証に失敗しました: %w", err)
			}

			if err := auth.SaveToken(token); err != nil {
				return fmt.Errorf("トークンの保存に失敗しました: %w", err)
			}

			fmt.Println("認証成功!")
			return nil
		},
	}
}
