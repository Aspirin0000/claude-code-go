package commands
import ("context"; "fmt"; "os"; "time")
type TouchCommand struct{*BaseCommand}
func NewTouchCommand() *TouchCommand { return &TouchCommand{BaseCommand: NewBaseCommand("touch", "Create empty file or update timestamp", CategoryFiles).WithAliases("create-file")}}
func (c *TouchCommand) Execute(ctx context.Context, args []string) error {
	if len(args) < 1 { return fmt.Errorf("usage: /touch <file>") }
	for _, f := range args {
		file, err := os.OpenFile(f, os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil { return err }
		file.Close()
		os.Chtimes(f, time.Now(), time.Now())
		fmt.Printf("✓ %s\n", f)
	}
	return nil
}
func init() { Register(NewTouchCommand()) }
