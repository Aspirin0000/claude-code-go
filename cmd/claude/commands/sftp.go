package commands

import (
	"context"
	"os"
	"os/exec"
)

type SftpCommand struct{ *BaseCommand }

func NewSftpCommand() *SftpCommand {
	return &SftpCommand{
		BaseCommand: NewBaseCommand("sftp", "Secure File Transfer Protocol", CategoryTools).
			WithHelp("Secure FTP client"),
	}
}

func (c *SftpCommand) Execute(ctx context.Context, args []string) error {
	cmd := exec.Command("sftp", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func init() { Register(NewSftpCommand()) }
