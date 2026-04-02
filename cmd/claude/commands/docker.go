package commands

import (
	"context"
	"fmt"
	"os"
	"os/exec"
)

type DockerCommand struct{ *BaseCommand }

func NewDockerCommand() *DockerCommand {
	return &DockerCommand{
		BaseCommand: NewBaseCommand(
			"docker",
			"Docker container platform",
			CategoryTools,
		).WithHelp(`Usage: /docker <command> [args]

Docker container management.

Common Commands:
  ps           List containers
  images       List images
  run          Run a container
  exec         Execute command in container
  build        Build an image
  pull         Pull an image
  push         Push an image
  stop         Stop a container
  rm           Remove a container
  rmi          Remove an image

Examples:
  /docker ps
  /docker images
  /docker run -it ubuntu
  /docker build -t myapp .
  /docker exec -it container_name bash`),
	}
}

func (c *DockerCommand) Execute(ctx context.Context, args []string) error {
	if len(args) < 1 {
		fmt.Println(c.Help())
		return nil
	}

	cmd := exec.CommandContext(ctx, "docker", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func init() { Register(NewDockerCommand()) }
