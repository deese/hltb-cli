package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
)

// version is overridden at build time via -ldflags "-X main.version=...".
var version = "dev"

const usage = `hltb - HowLongToBeat from your terminal

Usage:
  hltb [flags] <game name>

Flags:
  -n, --limit int   max results to show (default 3)
  -p, --page int    result page (default 1)
      --json        output raw JSON
  -v, --version     print version
`

type options struct {
	limit   int
	page    int
	asJSON  bool
	version bool
	query   string
}

func main() {
	os.Exit(run())
}

func run() int {
	opts, err := parseArgs(os.Args[1:])
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			fmt.Fprint(os.Stdout, usage)
			return 0
		}
		fmt.Fprintln(os.Stderr, errorStyle.Render("error:"), err)
		fmt.Fprint(os.Stderr, "\n"+usage)
		return 2
	}

	if opts.version {
		fmt.Println("hltb", version)
		return 0
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	res, err := NewClient().Search(ctx, opts.query, opts.page, opts.limit)
	if err != nil {
		fmt.Fprintln(os.Stderr, errorStyle.Render("error:"), err)
		return 1
	}

	if opts.asJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(res); err != nil {
			fmt.Fprintln(os.Stderr, errorStyle.Render("error:"), err)
			return 1
		}
		return 0
	}

	renderResults(os.Stdout, res, opts.query)
	return 0
}

// parseArgs accepts flags before or after the game name.
func parseArgs(args []string) (options, error) {
	var opts options
	fs := flag.NewFlagSet("hltb", flag.ContinueOnError)
	fs.SetOutput(discard{})
	fs.IntVar(&opts.limit, "limit", 3, "")
	fs.IntVar(&opts.limit, "n", 3, "")
	fs.IntVar(&opts.page, "page", 1, "")
	fs.IntVar(&opts.page, "p", 1, "")
	fs.BoolVar(&opts.asJSON, "json", false, "")
	fs.BoolVar(&opts.version, "version", false, "")
	fs.BoolVar(&opts.version, "v", false, "")

	var words []string
	for len(args) > 0 {
		if err := fs.Parse(args); err != nil {
			return opts, err
		}
		args = fs.Args()
		if len(args) == 0 {
			break
		}
		words = append(words, args[0])
		args = args[1:]
	}

	opts.query = strings.TrimSpace(strings.Join(words, " "))
	if opts.version {
		return opts, nil
	}
	if opts.query == "" {
		return opts, errors.New("missing game name")
	}
	if opts.limit < 1 || opts.limit > 20 {
		return opts, errors.New("--limit must be between 1 and 20")
	}
	if opts.page < 1 {
		return opts, errors.New("--page must be >= 1")
	}
	return opts, nil
}

type discard struct{}

func (discard) Write(p []byte) (int, error) { return len(p), nil }
