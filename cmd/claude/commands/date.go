package commands

import (
	"context"
	"fmt"
	"time"
)

type DateCommand struct{ *BaseCommand }

func NewDateCommand() *DateCommand {
	return &DateCommand{
		BaseCommand: NewBaseCommand("date", "Show current date and time", CategoryGeneral).
			WithHelp("Display current date and time"),
	}
}

func (c *DateCommand) Execute(ctx context.Context, args []string) error {
	now := time.Now()
	fmt.Println(now.Format("Mon Jan 2 15:04:05 MST 2006"))
	return nil
}

func init() { Register(NewDateCommand()) }
