package cmd

import (
	"context"
	"fmt"

	"github.com/qj0r9j0vc2/kko/internal/output"
	"github.com/qj0r9j0vc2/kko/internal/search"
	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search the web using Daum Search API",
	Long:  "Search for web pages, blog posts, and cafe articles.",
	Example: `  kko search "cosmos blockchain validator"
  kko search "golang error handling" --type blog --limit 3
  kko search "cosmos" --open 1`,
	Args: cobra.MinimumNArgs(1),
	RunE: runSearch,
}

var (
	searchType  string
	searchLimit int
	searchSort  string
	searchOpen  int
)

func init() {
	rootCmd.AddCommand(searchCmd)

	searchCmd.Flags().StringVarP(&searchType, "type", "t", "web", "search type: web, blog, cafe")
	searchCmd.Flags().IntVarP(&searchLimit, "limit", "l", 5, "max results")
	searchCmd.Flags().StringVarP(&searchSort, "sort", "s", "accuracy", "sort by: accuracy, recency")
	searchCmd.Flags().IntVarP(&searchOpen, "open", "o", 0, "open result N in browser")
}

func runSearch(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc := search.NewService(client)

	opts := search.SearchOptions{
		Query: args[0],
		Type:  searchType,
		Sort:  searchSort,
		Size:  searchLimit,
	}

	switch searchType {
	case "blog":
		return runSearchBlog(ctx, svc, opts)
	case "cafe":
		return runSearchCafe(ctx, svc, opts)
	default:
		return runSearchWeb(ctx, svc, opts)
	}
}

func runSearchWeb(ctx context.Context, svc *search.Service, opts search.SearchOptions) error {
	result, err := svc.Search(ctx, opts)
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

	fmt.Println()
	fmt.Println("  " + output.Header("WEB RESULTS"))
	fmt.Println()

	for i, doc := range result.Documents {
		fmt.Printf("  %s  %s\n", output.Label(fmt.Sprintf("%d", i+1)), search.StripHTML(doc.Title))
		fmt.Printf("     %s\n", output.Link(doc.URL))
		if doc.Contents != "" {
			content := search.StripHTML(doc.Contents)
			if len(content) > 80 {
				content = content[:80] + "..."
			}
			fmt.Printf("     %s\n", output.Muted(content))
		}
		fmt.Println()

		if searchOpen == i+1 {
			openURL(doc.URL)
			fmt.Printf("  %s\n\n", output.Success("Opened in browser"))
		}
	}

	return nil
}

func runSearchBlog(ctx context.Context, svc *search.Service, opts search.SearchOptions) error {
	result, err := svc.SearchBlog(ctx, opts)
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

	fmt.Println()
	fmt.Println("  " + output.Header("BLOG RESULTS"))
	fmt.Println()

	for i, doc := range result.Documents {
		fmt.Printf("  %s  %s\n", output.Label(fmt.Sprintf("%d", i+1)), search.StripHTML(doc.Title))
		fmt.Printf("     %s  %s\n", output.Link(doc.URL), output.Muted(doc.Blogname))
		if doc.Contents != "" {
			content := search.StripHTML(doc.Contents)
			if len(content) > 80 {
				content = content[:80] + "..."
			}
			fmt.Printf("     %s\n", output.Muted(content))
		}
		fmt.Println()

		if searchOpen == i+1 {
			openURL(doc.URL)
			fmt.Printf("  %s\n\n", output.Success("Opened in browser"))
		}
	}

	return nil
}

func runSearchCafe(ctx context.Context, svc *search.Service, opts search.SearchOptions) error {
	result, err := svc.SearchCafe(ctx, opts)
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

	fmt.Println()
	fmt.Println("  " + output.Header("CAFE RESULTS"))
	fmt.Println()

	for i, doc := range result.Documents {
		fmt.Printf("  %s  %s\n", output.Label(fmt.Sprintf("%d", i+1)), search.StripHTML(doc.Title))
		fmt.Printf("     %s  %s\n", output.Link(doc.URL), output.Muted(doc.Cafename))
		if doc.Contents != "" {
			content := search.StripHTML(doc.Contents)
			if len(content) > 80 {
				content = content[:80] + "..."
			}
			fmt.Printf("     %s\n", output.Muted(content))
		}
		fmt.Println()

		if searchOpen == i+1 {
			openURL(doc.URL)
			fmt.Printf("  %s\n\n", output.Success("Opened in browser"))
		}
	}

	return nil
}
