package cmd

import (
	"errors"
	"fmt"
	"sort"

	"github.com/charmbracelet/huh"
	"github.com/qj0r9j0vc2/kko/internal/local"
	"github.com/qj0r9j0vc2/kko/internal/output"
	"github.com/spf13/cobra"
)

var findCmd = &cobra.Command{
	Use:   "find [query]",
	Short: "Search for places using Kakao Local API",
	Long:  "Find places, businesses, and addresses near a location.",
	Example: `  kko find "starbucks" --near "hakdong"
  kko find --category cafe --near "hakdong" --radius 500
  kko find "starbucks" --near "hakdong" --json`,
	RunE: runFind,
}

var (
	findNear     string
	findRadius   int
	findCategory string
	findSort     string
	findLimit    int
	findOpen     bool
)

func init() {
	rootCmd.AddCommand(findCmd)

	findCmd.Flags().StringVarP(&findNear, "near", "n", "", `center point (name or "lat,lng")`)
	findCmd.Flags().IntVarP(&findRadius, "radius", "r", 2000, "search radius in meters")
	findCmd.Flags().StringVarP(&findCategory, "category", "c", "", "category (cafe, food, pharmacy...)")
	findCmd.Flags().StringVarP(&findSort, "sort", "s", "distance", "sort by: distance, accuracy")
	findCmd.Flags().IntVarP(&findLimit, "limit", "l", 5, "max results")
	findCmd.Flags().BoolVarP(&findOpen, "open", "o", false, "open first result in Kakao Map")
}

func runFind(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc := local.NewService(client)

	var x, y string
	nearLocation := findNear
	if nearLocation == "" {
		if cfg.DefaultLocation.Name != "" {
			nearLocation = cfg.DefaultLocation.Name
		}
		if cfg.DefaultLocation.Lat != 0 && cfg.DefaultLocation.Lng != 0 {
			x = fmt.Sprintf("%f", cfg.DefaultLocation.Lng)
			y = fmt.Sprintf("%f", cfg.DefaultLocation.Lat)
		}
	}
	if nearLocation != "" && x == "" {
		var err error
		x, y, err = svc.ResolveLocation(ctx, nearLocation)
		if err != nil {
			return fmt.Errorf("resolve --near location: %w", err)
		}
	}

	opts := local.SearchOptions{
		X:      x,
		Y:      y,
		Radius: findRadius,
		Sort:   findSort,
		Size:   findLimit,
	}

	var result *local.SearchResult
	var err error

	if findCategory != "" {
		opts.Category = findCategory
		result, err = svc.CategorySearch(ctx, opts)
	} else if len(args) == 0 {
		if !output.IsTerminal() {
			return usageError(cmd, "missing search query\n\n  Available categories: "+categoryList())
		}
		// Interactive category picker
		keys := make([]string, 0, len(local.CategoryCodes))
		for k := range local.CategoryCodes {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		options := make([]huh.Option[string], len(keys))
		for i, k := range keys {
			options[i] = huh.NewOption(k, k)
		}
		var picked string
		if err := huh.NewSelect[string]().
			Title("Pick a category").
			Options(options...).
			Value(&picked).
			Run(); err != nil {
			if errors.Is(err, huh.ErrUserAborted) {
				return nil
			}
			return err
		}
		opts.Category = picked
		result, err = svc.CategorySearch(ctx, opts)
	} else {
		opts.Query = args[0]
		result, err = svc.KeywordSearch(ctx, opts)
	}
	if err != nil {
		return err
	}

	if cfg.Output.Format == "json" {
		return output.PrintJSON(result)
	}

	if len(result.Documents) == 0 {
		fmt.Println(output.Muted("  No results found."))
		return nil
	}

	table := output.NewTable(
		output.Column{Header: "#", Width: 3, Align: output.AlignRight},
		output.Column{Header: "NAME"},
		output.Column{Header: "CATEGORY"},
		output.Column{Header: "DIST", Width: 6, Align: output.AlignRight},
		output.Column{Header: "ADDRESS"},
	)

	for i, place := range result.Documents {
		dist := place.Distance
		if dist != "" {
			dist = dist + "m"
		}
		table.AddRow(
			fmt.Sprintf("%d", i+1),
			place.PlaceName,
			place.CategoryName,
			dist,
			place.AddressName,
		)
	}

	fmt.Println()
	table.Render()
	fmt.Println()

	if findOpen && len(result.Documents) > 0 {
		mapURL := fmt.Sprintf("https://map.kakao.com/link/map/%s", result.Documents[0].ID)
		output.OpenURL(mapURL)
		fmt.Printf("  %s %s\n\n", output.Muted("Opened:"), output.Link(mapURL))
	}

	return nil
}
