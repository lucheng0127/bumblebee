package cli

import (
	"fmt"

	"github.com/lucheng0127/bumblebee/pkg/models"
	rCli "github.com/urfave/cli/v2"
	"xorm.io/xorm"
)

func addUser(cCtx *rCli.Context) error {
	dbClient, err := xorm.NewEngine("sqlite3", cCtx.String("database"))
	if err != nil {
		return err
	}

	username := cCtx.String("username")
	password := cCtx.String("password")
	user := models.NewUser(username, password)
	_, err = dbClient.Insert(user)
	if err != nil {
		return err
	}

	return nil
}

func listUser(cCtx *rCli.Context) error {
	dbClient, err := xorm.NewEngine("sqlite3", cCtx.String("database"))
	if err != nil {
		return err
	}

	var users []models.User
	dbClient.Find(&users)
	for _, user := range users {
		fmt.Println(user.Uuid, user.Name)
	}
	return nil
}

func NewUserCmd() *rCli.Command {
	return &rCli.Command{
		Name:  "user",
		Usage: "user client",
		Subcommands: []*rCli.Command{
			&rCli.Command{
				Name:   "add",
				Usage:  "add user",
				Action: addUser,
				Flags: []rCli.Flag{
					&rCli.StringFlag{
						Name:    "database",
						Aliases: []string{"d"},
						Usage:   "sqlite database file",
						Value:   "bumblebee.db",
					},
					&rCli.StringFlag{
						Name:     "username",
						Aliases:  []string{"u"},
						Usage:    "username",
						Required: true,
					},
					&rCli.StringFlag{
						Name:     "password",
						Aliases:  []string{"p"},
						Usage:    "password",
						Required: true,
					},
				},
			},
			&rCli.Command{
				Name:   "list",
				Usage:  "list user",
				Action: listUser,
				Flags: []rCli.Flag{
					&rCli.StringFlag{
						Name:    "database",
						Aliases: []string{"d"},
						Usage:   "sqlite database file",
						Value:   "bumblebee.db",
					},
				},
			},
		},
	}
}
