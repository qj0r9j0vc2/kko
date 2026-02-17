package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/qj0r9j0vc2/kko/internal/local"
	"github.com/qj0r9j0vc2/kko/internal/mobility"
	"github.com/qj0r9j0vc2/kko/internal/output"
	"github.com/spf13/cobra"
)

var routeCmd = &cobra.Command{
	Use:   "route [origin] [destination]",
	Short: "Get directions and fare estimates",
	Long:  "Get driving directions, estimated duration, and taxi fare between two locations.",
	Example: `  kko route "hakdong station" "euljiro 3-ga"
  kko route home office
  kko route home office --depart "08:30"`,
	Args: cobra.ExactArgs(2),
	RunE: runRoute,
}

var (
	routeDepart   string
	routePriority string
	routeAvoid    string
)

func init() {
	rootCmd.AddCommand(routeCmd)

	routeCmd.Flags().StringVarP(&routeDepart, "depart", "d", "", "departure time (HH:MM or YYYYMMDDHHMM)")
	routeCmd.Flags().StringVarP(&routePriority, "priority", "p", "RECOMMEND", "route priority: RECOMMEND, TIME, DISTANCE")
	routeCmd.Flags().StringVar(&routeAvoid, "avoid", "", "avoid: toll, motorway")
}

func runRoute(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	localSvc := local.NewService(client)
	mobSvc := mobility.NewService(client)

	origin := resolveAlias(args[0])
	dest := resolveAlias(args[1])

	ox, oy, err := localSvc.ResolveLocation(ctx, origin)
	if err != nil {
		return fmt.Errorf("resolve origin %q: %w", origin, err)
	}

	dx, dy, err := localSvc.ResolveLocation(ctx, dest)
	if err != nil {
		return fmt.Errorf("resolve destination %q: %w", dest, err)
	}

	var departTime string
	if routeDepart != "" {
		departTime, err = parseDepartTime(routeDepart)
		if err != nil {
			return err
		}
	}

	opts := mobility.DirectionsOptions{
		OriginX:    ox,
		OriginY:    oy,
		DestX:      dx,
		DestY:      dy,
		Priority:   routePriority,
		Avoid:      routeAvoid,
		DepartTime: departTime,
	}

	result, err := mobSvc.Directions(ctx, opts)
	if err != nil {
		return err
	}

	if cfg.Output.Format == "json" {
		return output.PrintJSON(result)
	}

	if len(result.Routes) == 0 {
		fmt.Println(output.Muted("  No route found."))
		return nil
	}

	route := result.Routes[0]
	if route.ResultCode != 0 {
		return fmt.Errorf("route error: %s", route.ResultMsg)
	}

	s := route.Summary

	fmt.Println()
	fmt.Printf("  %s %s -> %s\n", output.Header("Route:"), origin, dest)
	output.PrintDivider(nil, 45)

	distKm := float64(s.Distance) / 1000.0
	durMin := s.Duration / 60

	fmt.Printf("  %s   %.1f km\n", output.Label("Distance:"), distKm)
	fmt.Printf("  %s   %d min\n", output.Label("Duration:"), durMin)
	fmt.Printf("  %s  ~%s won\n", output.Label("Taxi fare:"), formatNumber(s.Fare.Taxi))
	fmt.Printf("  %s       %s won\n", output.Label("Toll:"), formatNumber(s.Fare.Toll))
	output.PrintDivider(nil, 45)

	if len(route.Sections) > 0 {
		var roads []string
		for _, section := range route.Sections {
			for _, road := range section.Roads {
				if road.Name != "" && !containsStr(roads, road.Name) {
					roads = append(roads, road.Name)
				}
			}
		}
		if len(roads) > 5 {
			roads = roads[:5]
		}
		if len(roads) > 0 {
			fmt.Printf("  %s %s\n", output.Muted("via"), strings.Join(roads, " -> "))
		}
	}

	fmt.Println()
	return nil
}

func resolveAlias(name string) string {
	if cfg != nil && cfg.Aliases != nil {
		if alias, ok := cfg.Aliases[name]; ok {
			return alias
		}
	}
	return name
}

func parseDepartTime(s string) (string, error) {
	if len(s) == 5 && s[2] == ':' {
		now := time.Now()
		t, err := time.Parse("15:04", s)
		if err != nil {
			return "", fmt.Errorf("invalid time format %q (use HH:MM)", s)
		}
		depart := time.Date(now.Year(), now.Month(), now.Day(), t.Hour(), t.Minute(), 0, 0, now.Location())
		if depart.Before(now) {
			depart = depart.Add(24 * time.Hour)
		}
		return depart.Format("200601021504"), nil
	}
	if len(s) == 12 {
		return s, nil
	}
	return "", fmt.Errorf("invalid departure time %q (use HH:MM or YYYYMMDDHHMM)", s)
}

func formatNumber(n int) string {
	s := fmt.Sprintf("%d", n)
	if len(s) <= 3 {
		return s
	}
	var result []byte
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			result = append(result, ',')
		}
		result = append(result, byte(c))
	}
	return string(result)
}

func containsStr(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}
