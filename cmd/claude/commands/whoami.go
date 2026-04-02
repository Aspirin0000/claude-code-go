package commands

import (
	"context"
	"fmt"
	"os"
)

type WhoamiCommand struct{ *BaseCommand }

func NewWhoamiCommand() *WhoamiCommand {
	return &WhoamiCommand{
		BaseCommand: NewBaseCommand("whoami", "Show current user", CategoryGeneral).
			WithHelp("Display the current user name"),
	}
}

func (c *WhoamiCommand) Execute(ctx context.Context, args []string) error {
	user, err := os.UserHomeDir()
	if err != nil {
		user = os.Getenv("USER")
		if user == "" {
			user = os.Getenv("USERNAME")
		}
	}
	if user == "" {
		user = "unknown"
	}
	fmt.Println(user)
	return nil
}

func init() { Register(NewWhoamiCommand()) }
