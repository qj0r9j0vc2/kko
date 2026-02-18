package cmd

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/qj0r9j0vc2/kko/internal/calendar"
	"github.com/qj0r9j0vc2/kko/internal/output"
	"github.com/spf13/cobra"
)

var calCmd = &cobra.Command{
	Use:   "cal",
	Short: "Manage Talk Calendar events and todos",
	Long:  "View, create, and delete events on your Kakao Talk Calendar. Requires OAuth login.",
}

var calTodayCmd = &cobra.Command{
	Use:   "today",
	Short: "Show today's events",
	RunE:  runCalToday,
}

var calTomorrowCmd = &cobra.Command{
	Use:   "tomorrow",
	Short: "Show tomorrow's events",
	RunE:  runCalTomorrow,
}

var calWeekCmd = &cobra.Command{
	Use:   "week",
	Short: "Show this week's events",
	RunE:  runCalWeek,
}

var calListCmd = &cobra.Command{
	Use:   "list",
	Short: "List events in a date range",
	RunE:  runCalList,
}

var calAddCmd = &cobra.Command{
	Use:   "add [title]",
	Short: "Create a new event",
	Example: `  kko cal add "Team standup" --at 14:00
  kko cal add "Lunch" --at 12:00 --dur 90
  kko cal add "Holiday" --date 2026-03-01 --all-day`,
	RunE: runCalAdd,
}

var calRmCmd = &cobra.Command{
	Use:   "rm [event_id]",
	Short: "Delete an event",
	Args:  cobra.ExactArgs(1),
	RunE:  runCalRm,
}

var calTodoCmd = &cobra.Command{
	Use:   "todo",
	Short: "List todos",
	RunE:  runCalTodoList,
}

var calTodoAddCmd = &cobra.Command{
	Use:   "add [task]",
	Short: "Add a todo",
	Args:  cobra.MinimumNArgs(1),
	RunE:  runCalTodoAdd,
}

var (
	calAt     string
	calDate   string
	calDur    int
	calDesc   string
	calAllDay bool
	calFrom   string
	calTo     string
)

func init() {
	rootCmd.AddCommand(calCmd)

	calCmd.AddCommand(calTodayCmd)
	calCmd.AddCommand(calTomorrowCmd)
	calCmd.AddCommand(calWeekCmd)
	calCmd.AddCommand(calListCmd)
	calCmd.AddCommand(calAddCmd)
	calCmd.AddCommand(calRmCmd)
	calCmd.AddCommand(calTodoCmd)
	calTodoCmd.AddCommand(calTodoAddCmd)

	calAddCmd.Flags().StringVarP(&calAt, "at", "a", "", "start time (HH:MM)")
	calAddCmd.Flags().StringVarP(&calDate, "date", "d", "", "date (YYYY-MM-DD), default today")
	calAddCmd.Flags().IntVar(&calDur, "dur", 60, "duration in minutes")
	calAddCmd.Flags().StringVar(&calDesc, "desc", "", "description")
	calAddCmd.Flags().BoolVar(&calAllDay, "all-day", false, "all-day event")

	calListCmd.Flags().StringVar(&calFrom, "from", "", "start date (YYYY-MM-DD)")
	calListCmd.Flags().StringVar(&calTo, "to", "", "end date (YYYY-MM-DD)")
}

func runCalToday(cmd *cobra.Command, _ []string) error {
	now := time.Now()
	from := startOfDay(now)
	to := endOfDay(now)
	return listEvents(cmd, from, to, "TODAY - "+now.Format("Mon, Jan 2, 2006"))
}

func runCalTomorrow(cmd *cobra.Command, _ []string) error {
	tmr := time.Now().Add(24 * time.Hour)
	from := startOfDay(tmr)
	to := endOfDay(tmr)
	return listEvents(cmd, from, to, "TOMORROW - "+tmr.Format("Mon, Jan 2, 2006"))
}

func runCalWeek(cmd *cobra.Command, _ []string) error {
	now := time.Now()
	weekday := int(now.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	monday := startOfDay(now.AddDate(0, 0, -(weekday - 1)))
	sunday := endOfDay(monday.AddDate(0, 0, 6))
	return listEvents(cmd, monday, sunday, "THIS WEEK")
}

func runCalList(cmd *cobra.Command, _ []string) error {
	var from, to time.Time
	var err error
	if calFrom != "" {
		from, err = time.Parse("2006-01-02", calFrom)
		if err != nil {
			return fmt.Errorf("invalid --from date: %w", err)
		}
	} else {
		from = startOfDay(time.Now())
	}
	if calTo != "" {
		to, err = time.Parse("2006-01-02", calTo)
		if err != nil {
			return fmt.Errorf("invalid --to date: %w", err)
		}
		to = endOfDay(to)
	} else {
		to = endOfDay(from.AddDate(0, 0, 7))
	}
	label := fmt.Sprintf("%s to %s", from.Format("Jan 2"), to.Format("Jan 2, 2006"))
	return listEvents(cmd, from, to, label)
}

func listEvents(cmd *cobra.Command, from, to time.Time, label string) error {
	ctx := cmd.Context()
	svc := calendar.NewService(client)

	events, err := svc.ListEvents(ctx, from, to)
	if err != nil {
		return err
	}

	if cfg.Output.Format == "json" {
		return output.PrintJSON(events)
	}

	fmt.Println()
	fmt.Println("  " + output.Header(label))
	output.PrintDivider(nil, 35)

	if len(events) == 0 {
		fmt.Println("  " + output.Muted("(no events)"))
	} else {
		for _, e := range events {
			timeStr := e.Time.StartAt.Format("15:04")
			if !e.Time.EndAt.IsZero() && e.Time.EndAt != e.Time.StartAt {
				timeStr += " - " + e.Time.EndAt.Format("15:04")
			}
			if e.Time.AllDay {
				timeStr = "all day"
			}
			fmt.Printf("  %s  %s\n", output.Label(timeStr), e.Title)
		}
	}

	fmt.Println()
	return nil
}

func runCalAdd(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc := calendar.NewService(client)

	title := strings.Join(args, " ")

	// Interactive form when no title given and running in a terminal
	if title == "" {
		if !output.IsTerminal() {
			return usageError(cmd, "missing event title")
		}
		var dateStr, timeStr, durStr, descStr string
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().Title("Title").Value(&title).Validate(func(s string) error {
					if strings.TrimSpace(s) == "" {
						return fmt.Errorf("title is required")
					}
					return nil
				}),
				huh.NewInput().Title("Date").Value(&dateStr).Placeholder("YYYY-MM-DD, default: today"),
				huh.NewInput().Title("Time").Value(&timeStr).Placeholder("HH:MM, leave empty for all-day"),
				huh.NewInput().Title("Duration (min)").Value(&durStr).Placeholder("60"),
				huh.NewInput().Title("Description").Value(&descStr).Placeholder("optional"),
			),
		)
		if err := form.Run(); err != nil {
			if errors.Is(err, huh.ErrUserAborted) {
				return nil
			}
			return err
		}
		if dateStr != "" {
			calDate = dateStr
		}
		if timeStr != "" {
			calAt = timeStr
		}
		if durStr != "" {
			d, err := strconv.Atoi(durStr)
			if err != nil {
				return fmt.Errorf("invalid duration: %w", err)
			}
			calDur = d
		}
		calDesc = descStr
	}

	date := time.Now()
	if calDate != "" {
		var err error
		date, err = time.Parse("2006-01-02", calDate)
		if err != nil {
			return fmt.Errorf("invalid --date: %w", err)
		}
	}

	var startAt, endAt time.Time
	if calAt != "" {
		t, err := time.Parse("15:04", calAt)
		if err != nil {
			return fmt.Errorf("invalid --at time: %w", err)
		}
		startAt = time.Date(date.Year(), date.Month(), date.Day(), t.Hour(), t.Minute(), 0, 0, date.Location())
		endAt = startAt.Add(time.Duration(calDur) * time.Minute)
	} else {
		startAt = startOfDay(date)
		endAt = endOfDay(date)
		calAllDay = true
	}

	req := &calendar.CreateEventRequest{
		Title: title,
		Time: &calendar.EventTime{
			StartAt:  startAt,
			EndAt:    endAt,
			TimeZone: "Asia/Seoul",
			AllDay:   calAllDay,
		},
		Description: calDesc,
	}

	event, err := svc.CreateEvent(ctx, req)
	if err != nil {
		return err
	}

	if cfg.Output.Format == "json" {
		return output.PrintJSON(event)
	}

	timeRange := startAt.Format("15:04") + "-" + endAt.Format("15:04")
	if calAllDay {
		timeRange = "all day"
	}
	fmt.Println()
	fmt.Printf("  %s Event created: %q at %s\n\n", output.Success("✓"), title, timeRange)
	return nil
}

func runCalRm(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc := calendar.NewService(client)

	if err := svc.DeleteEvent(ctx, args[0]); err != nil {
		return err
	}

	if cfg.Output.Format == "json" {
		return output.PrintJSON(map[string]string{"status": "deleted", "event_id": args[0]})
	}

	fmt.Println()
	fmt.Printf("  %s Event deleted.\n\n", output.Success("✓"))
	return nil
}

func runCalTodoList(cmd *cobra.Command, _ []string) error {
	ctx := cmd.Context()
	svc := calendar.NewService(client)

	todos, err := svc.ListTodos(ctx)
	if err != nil {
		return err
	}

	if cfg.Output.Format == "json" {
		return output.PrintJSON(todos)
	}

	fmt.Println()
	fmt.Println("  " + output.Header("TODOS"))
	output.PrintDivider(nil, 30)

	if len(todos) == 0 {
		fmt.Println("  " + output.Muted("(no todos)"))
	} else {
		for _, t := range todos {
			status := "[ ]"
			if t.Completed {
				status = "[x]"
			}
			fmt.Printf("  %s %s\n", output.Label(status), t.Content)
		}
	}
	fmt.Println()
	return nil
}

func runCalTodoAdd(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc := calendar.NewService(client)

	content := args[0]
	todo, err := svc.CreateTodo(ctx, content, nil)
	if err != nil {
		return err
	}

	if cfg.Output.Format == "json" {
		return output.PrintJSON(todo)
	}

	fmt.Println()
	fmt.Printf("  %s Todo added: %q\n\n", output.Success("✓"), content)
	return nil
}

func startOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

func endOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 0, t.Location())
}
