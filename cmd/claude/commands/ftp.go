package commands

import (
	"context"
	"os"
	"os/exec"
)

type FtpCommand struct{ *BaseCommand }

func NewFtpCommand() *FtpCommand {
	return &FtpCommand{
		BaseCommand: NewBaseCommand("ftp", "File Transfer Protocol client", CategoryTools).
			WithHelp("FTP client for file transfers"),
	}
}

func (c *FtpCommand) Execute(ctx context.Context, args []string) error {
	cmd := exec.Command("ftp", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func init() { Register(NewFtpCommand()) }
