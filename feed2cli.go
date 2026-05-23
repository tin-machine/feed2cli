package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mmcdole/gofeed"
	"golang.org/x/term"
)

type cliConfig struct {
	isDebug           bool
	createSymlinks    bool
	operation         string
	filterType        string
	inputFormat       string
	configPath        string
	pipelineStages    []FeedItemStage
	explain           bool
	enrichTypes       []string
	includeKeyword    []string
	excludeKeyword    []string
	minKeywordScore   int
	includeDomain     []string
	excludeDomain     []string
	since             time.Duration
	favUsers          []string
	rankBy            string
	minHotness        float64
	feedURLs          []string
	slackChannel      string
	slackDryRun       bool
	slackSkipValidate bool
	stateBackend      string
	statePath         string
	digestWindow      time.Duration
	digestTitle       string
}

func parseArgs(args []string, stderr io.Writer) (cliConfig, error) {
	var cfg cliConfig
	fs := flag.NewFlagSet(commandName(args), flag.ContinueOnError)
	fs.SetOutput(stderr)
	fs.BoolVar(&cfg.isDebug, "d", false, "Debug output")
	fs.BoolVar(&cfg.createSymlinks, "s", false, "Create symbolic links")
	fs.StringVar(&cfg.operation, "o", "", "Operation: merge, diff, atom, slack, hatena, digest, lint, json, jsonl")
	fs.StringVar(&cfg.filterType, "f", "", "Filter to apply: hatena_bookmark")
	fs.StringVar(&cfg.inputFormat, "input", "feed", "Input format: feed, jsonl")
	fs.StringVar(&cfg.configPath, "config", "", "Pipeline config JSON file")
	fs.BoolVar(&cfg.explain, "explain", false, "Print item-level explain JSONL and do not execute output")
	var enrichTypes stringListFlag
	fs.Var(&enrichTypes, "enrich", "Enrichment to apply: source_label, tag, ogp, content, summary. Can be specified multiple times")
	var includeKeyword stringListFlag
	fs.Var(&includeKeyword, "include-keyword", "Keep items containing keyword. Can be specified multiple times")
	var excludeKeyword stringListFlag
	fs.Var(&excludeKeyword, "exclude-keyword", "Drop items containing keyword. Can be specified multiple times")
	fs.IntVar(&cfg.minKeywordScore, "min-keyword-score", 0, "Minimum include keyword score")
	var includeDomain stringListFlag
	fs.Var(&includeDomain, "include-domain", "Keep items from domain. Can be specified multiple times")
	var excludeDomain stringListFlag
	fs.Var(&excludeDomain, "exclude-domain", "Drop items from domain. Can be specified multiple times")
	fs.DurationVar(&cfg.since, "since", 0, "Keep items newer than this duration, e.g. 24h")
	var favUsers stringListFlag
	fs.Var(&favUsers, "fav-user", "Keep items bookmarked by Hatena user. Can be specified multiple times")
	fs.StringVar(&cfg.rankBy, "rank", "", "Rank items by: hotness, published")
	fs.Float64Var(&cfg.minHotness, "min-hotness", 0, "Minimum hotness score")
	var feedURLs stringListFlag
	fs.Var(&feedURLs, "url", "Feed URL to fetch. Can be specified multiple times")
	fs.StringVar(&cfg.slackChannel, "slack-channel", "", "Slack channel ID/name override")
	fs.BoolVar(&cfg.slackDryRun, "slack-dry-run", false, "Print Slack post plan without sending")
	fs.BoolVar(&cfg.slackSkipValidate, "slack-skip-validate", false, "Skip Slack auth/channel validation before posting")
	fs.StringVar(&cfg.stateBackend, "state-backend", envDefault("FEED2CLI_STATE_BACKEND", "json"), "State backend: json, sqlite")
	fs.StringVar(&cfg.statePath, "state-path", os.Getenv("FEED2CLI_STATE_PATH"), "State file path")
	fs.DurationVar(&cfg.digestWindow, "digest-window", 24*time.Hour, "Digest time window. Use 0 to include all items")
	fs.StringVar(&cfg.digestTitle, "digest-title", "feed2cli digest", "Digest title")

	parseArgs := []string{}
	if len(args) > 1 {
		parseArgs = args[1:]
	}
	if err := fs.Parse(parseArgs); err != nil {
		return cliConfig{}, err
	}
	cfg.feedURLs = []string(feedURLs)
	cfg.enrichTypes = []string(enrichTypes)
	cfg.includeKeyword = []string(includeKeyword)
	cfg.excludeKeyword = []string(excludeKeyword)
	cfg.includeDomain = []string(includeDomain)
	cfg.excludeDomain = []string(excludeDomain)
	cfg.favUsers = []string(favUsers)
	cfg.inputFormat = strings.ToLower(strings.TrimSpace(cfg.inputFormat))
	if cfg.inputFormat == "" {
		cfg.inputFormat = "feed"
	}
	cfg.stateBackend = strings.ToLower(strings.TrimSpace(cfg.stateBackend))
	if cfg.stateBackend == "" {
		cfg.stateBackend = "json"
	}
	if cfg.statePath == "" {
		if cfg.stateBackend == "sqlite" {
			cfg.statePath = "hatena_state.sqlite"
		} else {
			cfg.statePath = stateFilePath
		}
	}
	return cfg, nil
}

type stringListFlag []string

func (f *stringListFlag) String() string {
	if f == nil {
		return ""
	}
	return strings.Join(*f, ",")
}

func (f *stringListFlag) Set(value string) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return fmt.Errorf("empty value")
	}
	*f = append(*f, value)
	return nil
}

func envDefault(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}

func commandName(args []string) string {
	if len(args) == 0 || args[0] == "" {
		return "feed2cli"
	}
	return filepath.Base(args[0])
}

func printDebugArgs(w io.Writer, args []string) {
	for i, v := range args {
		fmt.Fprintf(w, "args[%d] -> %s\n", i, v)
	}
}

func createSymlinksIfNeeded() error {
	for _, link := range []string{"mergeRss", "diffRss", "slackRss", "hatenaRss"} {
		if err := os.Symlink("feed2cli", link); err != nil && !os.IsExist(err) {
			return err
		}
	}
	return nil
}

func applyFilter(filterType string, feeds []*gofeed.Feed) ([]*FilteredItem, error) {
	var f Filter
	switch filterType {
	case "hatena_bookmark":
		f = &HatenaBookmarkFilter{}
	default:
		// フィルタが指定されていない、または未対応の場合は、型変換のみ行う
		return convertToFilteredItems(feeds), nil
	}

	allItems := []*gofeed.Item{}
	for _, feed := range feeds {
		allItems = append(allItems, feed.Items...)
	}

	return f.Apply(allItems)
}

func dispatchOperation(cfg cliConfig, cmd string, data interface{}, stdout, stderr io.Writer) int {
	op := operationName(cfg.operation, cmd)

	switch op {
	case "merge":
		feeds := convertToFeeds(data)
		merged := Merge(feeds)
		if err := OutputStandardTo(stdout, merged, time.Now()); err != nil {
			fmt.Fprintf(stderr, "RSS出力に失敗しました: %v\n", err)
			return 1
		}
	case "diff":
		feeds := convertToFeeds(data)
		diffed := Diff(feeds)
		if err := OutputStandardTo(stdout, diffed, time.Now()); err != nil {
			fmt.Fprintf(stderr, "RSS出力に失敗しました: %v\n", err)
			return 1
		}
	case "slack":
		if err := OutputSlackWithOptions(data, slackOutputOptions{
			Channel:               cfg.slackChannel,
			DryRun:                cfg.slackDryRun,
			SkipChannelValidation: cfg.slackSkipValidate,
			DryRunWriter:          stdout,
		}); err != nil {
			fmt.Fprintf(stderr, "%v\n", err)
			return 1
		}
	case "hatena":
		if items, ok := data.([]*FilteredItem); ok {
			if err := OutputHatenaToSlackWithOptions(items, hatenaOutputOptions{
				Channel:               cfg.slackChannel,
				DryRun:                cfg.slackDryRun,
				SkipChannelValidation: cfg.slackSkipValidate,
				DryRunWriter:          stdout,
				StateBackend:          cfg.stateBackend,
				StatePath:             cfg.statePath,
			}); err != nil {
				fmt.Fprintf(stderr, "%v\n", err)
				return 1
			}
		} else {
			fmt.Fprintln(stderr, "hatena操作にはフィルタリングされたデータが必要です。-f hatena_bookmark を使用してください。")
			return 1
		}
	case "json":
		if err := OutputJSONTo(stdout, data); err != nil {
			fmt.Fprintf(stderr, "JSON出力に失敗しました: %v\n", err)
			return 1
		}
	case "jsonl":
		if err := OutputJSONLTo(stdout, data); err != nil {
			fmt.Fprintf(stderr, "JSONL出力に失敗しました: %v\n", err)
			return 1
		}
	case "atom":
		if err := OutputAtomTo(stdout, data, time.Now()); err != nil {
			fmt.Fprintf(stderr, "Atom出力に失敗しました: %v\n", err)
			return 1
		}
	case "digest":
		if err := OutputDigestMarkdownTo(stdout, data, DigestOptions{
			Title:  cfg.digestTitle,
			Window: cfg.digestWindow,
			Now:    time.Now(),
		}); err != nil {
			fmt.Fprintf(stderr, "digest出力に失敗しました: %v\n", err)
			return 1
		}
	default:
		if err := OutputStandardTo(stdout, data, time.Now()); err != nil {
			fmt.Fprintf(stderr, "RSS出力に失敗しました: %v\n", err)
			return 1
		}
	}
	return 0
}

func main() {
	os.Exit(run(os.Args, os.Stdin, os.Stdout, os.Stderr, term.IsTerminal(0)))
}

func run(args []string, stdin io.Reader, stdout, stderr io.Writer, isTerminal bool) int {
	cfg, err := parseArgs(args, stderr)
	if err != nil {
		return 2
	}
	if cfg.isDebug {
		printDebugArgs(stderr, args)
	}
	if cfg.configPath != "" {
		pipeline, err := LoadPipelineConfig(cfg.configPath)
		if err != nil {
			fmt.Fprintf(stderr, "pipeline configの読み込みに失敗しました: %v\n", err)
			return 1
		}
		cfg, err = applyPipelineConfig(cfg, pipeline)
		if err != nil {
			fmt.Fprintf(stderr, "pipeline configの適用に失敗しました: %v\n", err)
			return 1
		}
	}
	if isTerminal {
		fmt.Fprintln(stderr, "パイプ無し(FD値0)")
		if cfg.createSymlinks {
			if err := createSymlinksIfNeeded(); err != nil {
				fmt.Fprintf(stderr, "symlink作成に失敗しました: %v\n", err)
				return 1
			}
		}
		if cfg.operation == "" && cfg.filterType == "" && !cfg.explain && !cfg.hasItemStages() && len(cfg.feedURLs) == 0 {
			fmt.Fprintln(stderr, "操作またはフィルタを指定してください: -o <operation> | -f <filter>")
			return 0
		}
		if len(cfg.feedURLs) == 0 {
			return 0
		}
	}

	cmd := commandName(args)
	if operationName(cfg.operation, cmd) == "lint" {
		if cfg.inputFormat != "feed" {
			fmt.Fprintln(stderr, "lint操作は -input feed のみ対応しています。")
			return 1
		}
		var result FeedLintResult
		if !isTerminal {
			result = LintFeedsFromReader(stdin)
		}
		if len(cfg.feedURLs) > 0 {
			result = MergeFeedLintResults(result, LintFeedURLs(cfg.feedURLs))
		}
		if err := OutputFeedLintTo(stdout, result); err != nil {
			fmt.Fprintf(stderr, "feed lint出力に失敗しました: %v\n", err)
			return 1
		}
		if result.Invalid > 0 {
			return 1
		}
		return 0
	}

	dataToDispatch, feeds, err := readInputData(cfg, stdin, isTerminal)
	if err != nil {
		fmt.Fprintf(stderr, "%v\n", err)
		return 1
	}

	if cfg.filterType != "" {
		if cfg.inputFormat != "feed" {
			fmt.Fprintln(stderr, "-f は -input feed のみ対応しています。JSONL入力では -enrich や filter/rank flags を使ってください。")
			return 1
		}
		filteredItems, err := applyFilter(cfg.filterType, feeds)
		if err != nil {
			fmt.Fprintf(stderr, "フィルタの適用に失敗しました: %v\n", err)
			return 1
		}
		dataToDispatch = filteredItems
	}
	if cfg.explain {
		if err := OutputExplainJSONLTo(stdout, dataToDispatch, cfg, cmd); err != nil {
			fmt.Fprintf(stderr, "explain出力に失敗しました: %v\n", err)
			return 1
		}
		return 0
	}
	if len(cfg.pipelineStages) > 0 {
		items, err := RunFeedItemStages(context.Background(), FeedItemsFromData(dataToDispatch), cfg.pipelineStages...)
		if err != nil {
			fmt.Fprintf(stderr, "pipeline config stageの適用に失敗しました: %v\n", err)
			return 1
		}
		dataToDispatch = items
	}
	if len(cfg.enrichTypes) > 0 {
		enrichedItems, err := applyEnrichments(cfg.enrichTypes, dataToDispatch)
		if err != nil {
			fmt.Fprintf(stderr, "enrichの適用に失敗しました: %v\n", err)
			return 1
		}
		dataToDispatch = enrichedItems
	}
	if cfg.hasFilterRankStages() {
		items, err := applyFilterRankStages(cfg, dataToDispatch)
		if err != nil {
			fmt.Fprintf(stderr, "filter/rankの適用に失敗しました: %v\n", err)
			return 1
		}
		dataToDispatch = items
	}

	return dispatchOperation(cfg, cmd, dataToDispatch, stdout, stderr)
}

func readInputData(cfg cliConfig, stdin io.Reader, isTerminal bool) (interface{}, []*gofeed.Feed, error) {
	switch cfg.inputFormat {
	case "feed", "rss", "atom":
		var feeds []*gofeed.Feed
		if !isTerminal {
			feeds = readFrom(stdin)
		}
		if len(cfg.feedURLs) > 0 {
			urlFeeds, err := fetchFeedsFromURLs(cfg.feedURLs)
			if err != nil {
				return nil, nil, err
			}
			feeds = append(feeds, urlFeeds...)
		}
		return feeds, feeds, nil
	case "jsonl":
		var items []FeedItem
		if !isTerminal {
			readItems, err := ReadJSONLItems(stdin)
			if err != nil {
				return nil, nil, err
			}
			items = append(items, readItems...)
		}
		if len(cfg.feedURLs) > 0 {
			urlFeeds, err := fetchFeedsFromURLs(cfg.feedURLs)
			if err != nil {
				return nil, nil, err
			}
			items = append(items, FeedItemsFromData(urlFeeds)...)
		}
		return items, FeedItemsToFeeds(items), nil
	default:
		return nil, nil, fmt.Errorf("unsupported input format %q", cfg.inputFormat)
	}
}

func FeedItemsToFeeds(items []FeedItem) []*gofeed.Feed {
	return []*gofeed.Feed{FeedFromItems(items)}
}

func (cfg cliConfig) hasItemStages() bool {
	return len(cfg.pipelineStages) > 0 || len(cfg.enrichTypes) > 0 || cfg.hasFilterRankStages()
}

func (cfg cliConfig) hasFilterRankStages() bool {
	return len(cfg.includeKeyword) > 0 ||
		len(cfg.excludeKeyword) > 0 ||
		cfg.minKeywordScore > 0 ||
		len(cfg.includeDomain) > 0 ||
		len(cfg.excludeDomain) > 0 ||
		cfg.since > 0 ||
		len(cfg.favUsers) > 0 ||
		cfg.rankBy != "" ||
		cfg.minHotness > 0
}

func applyFilterRankStages(cfg cliConfig, data interface{}) ([]FeedItem, error) {
	items := FeedItemsFromData(data)
	stages := filterRankStages(cfg)
	return RunFeedItemStages(context.Background(), items, stages...)
}

func filterRankStages(cfg cliConfig) []FeedItemStage {
	var stages []FeedItemStage
	if len(cfg.includeKeyword) > 0 || len(cfg.excludeKeyword) > 0 || cfg.minKeywordScore > 0 {
		stages = append(stages, KeywordFilterStage{
			Include:  cfg.includeKeyword,
			Exclude:  cfg.excludeKeyword,
			MinScore: cfg.minKeywordScore,
		})
	}
	if len(cfg.includeDomain) > 0 || len(cfg.excludeDomain) > 0 {
		stages = append(stages, DomainFilterStage{
			Include: cfg.includeDomain,
			Exclude: cfg.excludeDomain,
		})
	}
	if cfg.since > 0 {
		stages = append(stages, TimeWindowFilterStage{Since: cfg.since})
	}
	rankBy := strings.ToLower(strings.TrimSpace(cfg.rankBy))
	if rankBy == "hotness" || cfg.minHotness > 0 {
		stages = append(stages, HotnessScoreStage{})
	}
	if cfg.minHotness > 0 {
		stages = append(stages, MinHotnessStage{Min: cfg.minHotness})
	}
	if len(cfg.favUsers) > 0 {
		stages = append(stages, FavUserStage{Users: cfg.favUsers})
	}
	if rankBy != "" {
		stages = append(stages, RankStage{By: rankBy})
	}
	return stages
}

func applyEnrichments(enrichTypes []string, data interface{}) ([]FeedItem, error) {
	items := FeedItemsFromData(data)
	stages, err := enrichStages(enrichTypes)
	if err != nil {
		return nil, err
	}
	return RunFeedItemStages(context.Background(), items, stages...)
}

func enrichStages(enrichTypes []string) ([]FeedItemStage, error) {
	stages := make([]FeedItemStage, 0, len(enrichTypes))
	for _, enrichType := range enrichTypes {
		switch strings.ToLower(strings.TrimSpace(enrichType)) {
		case "source_label", "source-label", "source":
			stages = append(stages, SourceLabelStage{})
		case "tag", "tags":
			stages = append(stages, TagEnrichStage{})
		case "ogp":
			stages = append(stages, OGPEnrichStage{})
		case "content", "readability":
			stages = append(stages, ContentFetchStage{})
		case "summary", "local_summary", "local-llm-summary":
			stages = append(stages, SummaryStage{})
		default:
			return nil, fmt.Errorf("unsupported enrich %q", enrichType)
		}
	}
	return stages, nil
}

func operationName(operation, cmd string) string {
	if operation != "" {
		return operation
	}
	return strings.TrimSuffix(cmd, "Rss")
}

func convertToFeeds(data interface{}) []*gofeed.Feed {
	if feeds, ok := data.([]*gofeed.Feed); ok {
		return feeds
	}

	if items, ok := data.([]*FilteredItem); ok {
		return []*gofeed.Feed{FeedFromItems(FeedItemsFromFilteredItems(items))}
	}

	if items, ok := data.([]*gofeed.Item); ok {
		return []*gofeed.Feed{{Items: items}}
	}

	return nil
}
