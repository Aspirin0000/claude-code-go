package commands

import (
	"context"
	"fmt"
	"os"
	"os/exec"
)

type KubectlCommand struct{ *BaseCommand }

func NewKubectlCommand() *KubectlCommand {
	return &KubectlCommand{
		BaseCommand: NewBaseCommand(
			"kubectl",
			"Kubernetes command-line tool",
			CategoryTools,
		).WithHelp(`Usage: /kubectl <command> [args]

Kubernetes cluster management.

Common Commands:
  get          Get resources
  describe     Show details
  apply        Apply configuration
  delete       Delete resources
  logs         Print pod logs
  exec         Execute command in pod
  port-forward Forward ports
  create       Create resources

Examples:
  /kubectl get pods
  /kubectl get nodes
  /kubectl logs pod-name
  /kubectl apply -f config.yaml
  /kubectl exec -it pod-name -- bash`),
	}
}

func (c *KubectlCommand) Execute(ctx context.Context, args []string) error {
	if len(args) < 1 {
		fmt.Println(c.Help())
		return nil
	}

	cmd := exec.CommandContext(ctx, "kubectl", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func init() { Register(NewKubectlCommand()) }
