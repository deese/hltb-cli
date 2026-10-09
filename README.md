# hltb-cli

Look up [HowLongToBeat](https://howlongtobeat.com) playtimes from your terminal. Single static binary, written in Go.

## Usage

```
hltb [flags] <game name>

  -n, --limit int   max results to show (default 3)
  -p, --page int    result page (default 1)
      --json        output raw JSON
  -v, --version     print version
```

```
hltb elden ring
hltb -n 5 zelda
hltb --json hollow knight
```

Each result shows the main story, main + extras, completionist and all-styles times (plus co-op and versus when available), the review score, and how many people are playing, have it in their backlog or have beaten it. Colors respect `NO_COLOR` and are disabled when output is not a terminal.

## Install

Download the archive for your platform from the [releases page](https://github.com/deese/hltb-cli/releases), verify it against `checksums.txt`, and put `hltb` somewhere in your `PATH`.

Or build from source (Go 1.25+):

```
CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o bin/hltb .
```

## Disclaimer

HowLongToBeat has no public API. This tool uses the same undocumented endpoints as the website, which can change or break at any time and may be rate limited. It is not affiliated with or endorsed by HowLongToBeat. Check their terms of use before relying on it beyond personal use.

## License

[MIT](LICENSE)
