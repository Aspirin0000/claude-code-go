package commands

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"
)

type EnvCommand struct{ *BaseCommand }

func NewEnvCommand() *EnvCommand {
	return &EnvCommand{
		BaseCommand: NewBaseCommand("env", "Show environment variables", CategoryGeneral).
			WithHelp("Display all environment variables or search for specific ones"),
	}
}

func (c *EnvCommand) Execute(ctx context.Context, args []string) error {
	env := os.Environ()
	sort.Strings(env)

	if len(args) > 0 {
		search := strings.ToLower(args[0])
		for _, e := range env {
			if strings.Contains(strings.ToLower(e), search) {
				fmt.Println(e)
			}
		}
	} else {
		for _, e := range env {
			fmt.Println(e)
		}
	}
	return nil
}

func init() { Register(NewEnvCommand()) }
