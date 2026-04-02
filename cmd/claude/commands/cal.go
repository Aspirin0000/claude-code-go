package commands

import (
	"context"
	"fmt"
	"time"
)

type CalCommand struct{ *BaseCommand }

func NewCalCommand() *CalCommand {
	return &CalCommand{
		BaseCommand: NewBaseCommand("cal", "Display calendar", CategoryGeneral).
			WithHelp("Show calendar for current month or specified month/year"),
	}
}

func (c *CalCommand) Execute(ctx context.Context, args []string) error {
	now := time.Now()
	year, month, _ := now.Date()

	// Simple calendar display
	fmt.Printf("   %s %d\n", month.String(), year)
	fmt.Println("Su Mo Tu We Th Fr Sa")

	// Calculate first day of month
	firstDay := time.Date(year, month, 1, 0, 0, 0, 0, time.Local)
	startWeekday := int(firstDay.Weekday())

	// Print leading spaces
	for i := 0; i < startWeekday; i++ {
		fmt.Print("   ")
	}

	// Print days
	daysInMonth := 31
	switch month {
	case time.April, time.June, time.September, time.November:
		daysInMonth = 30
	case time.February:
		if year%4 == 0 && (year%100 != 0 || year%400 == 0) {
			daysInMonth = 29
		} else {
			daysInMonth = 28
		}
	}

	for day := 1; day <= daysInMonth; day++ {
		fmt.Printf("%2d ", day)
		if (startWeekday+day)%7 == 0 {
			fmt.Println()
		}
	}
	fmt.Println()
	return nil
}

func init() { Register(NewCalCommand()) }
