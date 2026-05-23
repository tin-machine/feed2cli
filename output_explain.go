package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
	"time"
)

const explainJSONLSchemaVersion = "feed2cli.explain.v1"

type ExplainRecord struct {
	SchemaVersion string               `json:"schema_version"`
	Output        string               `json:"output"`
	Kept          bool                 `json:"kept"`
	Reasons       []string             `json:"reasons,omitempty"`
	Stages        []ExplainStageRecord `json:"stages,omitempty"`
	Item          FeedItem             `json:"item"`
}

type ExplainStageRecord struct {
	Stage   string   `json:"stage"`
	Kept    bool     `json:"kept"`
	Reasons []string `json:"reasons,omitempty"`
}

type explainItemState struct {
	RecordIndex int
	Item        FeedItem
}

func OutputExplainJSONLTo(w io.Writer, data interface{}, cfg cliConfig, cmd string) error {
	items, err := outputFeedItems(data)
	if err != nil {
		return err
	}
	records := ExplainItems(items, cfg, explainOutputName(cfg, cmd), time.Now())
	encoder := json.NewEncoder(w)
	for _, record := range records {
		if err := encoder.Encode(record); err != nil {
			return err
		}
	}
	return nil
}

func ExplainItems(items []FeedItem, cfg cliConfig, output string, now time.Time) []ExplainRecord {
	records := make([]ExplainRecord, 0, len(items))
	states := make([]explainItemState, 0, len(items))
	for i, item := range items {
		records = append(records, ExplainRecord{
			SchemaVersion: explainJSONLSchemaVersion,
			Output:        output,
			Kept:          true,
			Item:          item,
		})
		states = append(states, explainItemState{RecordIndex: i, Item: item})
	}

	stages, stageErrors := explainStages(cfg)
	for _, err := range stageErrors {
		for i := range records {
			records[i].Reasons = append(records[i].Reasons, "explain: "+err.Error())
		}
	}
	if len(stages) == 0 {
		for i := range records {
			records[i].Reasons = append(records[i].Reasons, "no filter/rank criteria")
		}
		return records
	}

	for _, stage := range stages {
		states = explainStage(stage, states, records, now)
	}
	for i := range records {
		records[i].Kept = false
	}
	for _, state := range states {
		records[state.RecordIndex].Kept = true
		records[state.RecordIndex].Item = state.Item
	}
	for i := range records {
		if len(records[i].Reasons) == 0 {
			records[i].Reasons = flattenStageReasons(records[i].Stages)
		}
	}
	return records
}

func explainStages(cfg cliConfig) ([]FeedItemStage, []error) {
	var stages []FeedItemStage
	var errs []error
	stages = append(stages, cfg.pipelineStages...)
	if len(cfg.enrichTypes) > 0 {
		enrich, err := enrichStages(cfg.enrichTypes)
		if err != nil {
			errs = append(errs, err)
		} else {
			stages = append(stages, enrich...)
		}
	}
	stages = append(stages, filterRankStages(cfg)...)
	return stages, errs
}

func explainStage(stage FeedItemStage, states []explainItemState, records []ExplainRecord, now time.Time) []explainItemState {
	switch s := stage.(type) {
	case NormalizeStage:
		return explainTransformStage(stage.Name(), states, records, "normalize: set normalized_url/id", func(items []FeedItem) ([]FeedItem, error) {
			return s.Apply(context.Background(), items)
		})
	case MergeStage:
		return explainMergeStage(states, records)
	case KeywordFilterStage:
		return explainFilterStage(stage.Name(), states, records, func(item FeedItem) (FeedItem, bool, []string) {
			return explainKeywordStage(item, s)
		})
	case DomainFilterStage:
		return explainFilterStage(stage.Name(), states, records, func(item FeedItem) (FeedItem, bool, []string) {
			return explainDomainStage(item, s)
		})
	case TimeWindowFilterStage:
		if s.Now.IsZero() {
			s.Now = now
		}
		return explainFilterStage(stage.Name(), states, records, func(item FeedItem) (FeedItem, bool, []string) {
			return explainTimeWindowStage(item, s)
		})
	case HotnessScoreStage:
		if s.Now.IsZero() {
			s.Now = now
		}
		return explainFilterStage(stage.Name(), states, records, func(item FeedItem) (FeedItem, bool, []string) {
			score := hotnessScore(item, s.Now)
			item.Metadata = cloneMetadata(item.Metadata)
			item.Metadata["hotness_score"] = formatScore(score)
			return item, true, []string{"hotness_score: score=" + formatScore(score)}
		})
	case MinHotnessStage:
		return explainFilterStage(stage.Name(), states, records, func(item FeedItem) (FeedItem, bool, []string) {
			return explainMinHotnessStage(item, s)
		})
	case FavUserStage:
		return explainFilterStage(stage.Name(), states, records, func(item FeedItem) (FeedItem, bool, []string) {
			return explainFavUserStage(item, s)
		})
	case RankStage:
		return explainRankStage(s, states, records)
	case SourceLabelStage:
		return explainTransformStage(stage.Name(), states, records, "source_label: set source label when available", func(items []FeedItem) ([]FeedItem, error) {
			return s.Apply(context.Background(), items)
		})
	case TagEnrichStage:
		return explainTransformStage(stage.Name(), states, records, "tag: merge Hatena comment tags into categories", func(items []FeedItem) ([]FeedItem, error) {
			return s.Apply(context.Background(), items)
		})
	case SummaryStage:
		return explainTransformStage(stage.Name(), states, records, "summary: local extractive summary without external API", func(items []FeedItem) ([]FeedItem, error) {
			return s.Apply(context.Background(), items)
		})
	case OGPEnrichStage:
		return explainDryRunStage(stage.Name(), states, records, "ogp: would fetch article URL and read OGP metadata")
	case ContentFetchStage:
		return explainDryRunStage(stage.Name(), states, records, "content: would fetch article URL and extract readable content")
	case HatenaBookmarkStage:
		return explainDryRunStage(stage.Name(), states, records, "hatena_bookmark: would call Hatena APIs")
	case ExternalCommandStage:
		reason := "plugin: dry-run command=" + s.Command
		if len(s.Args) > 0 {
			reason += " args=" + strings.Join(s.Args, " ")
		}
		if s.Timeout > 0 {
			reason += " timeout=" + s.Timeout.String()
		}
		return explainDryRunStage(stage.Name(), states, records, reason)
	default:
		return explainDryRunStage(stage.Name(), states, records, "stage: explain is not implemented; assuming kept")
	}
}

func explainFilterStage(stageName string, states []explainItemState, records []ExplainRecord, eval func(FeedItem) (FeedItem, bool, []string)) []explainItemState {
	out := make([]explainItemState, 0, len(states))
	for _, state := range states {
		item, kept, reasons := eval(state.Item)
		appendExplainStage(records, state.RecordIndex, stageName, kept, reasons)
		if kept {
			state.Item = item
			out = append(out, state)
		}
	}
	return out
}

func explainTransformStage(stageName string, states []explainItemState, records []ExplainRecord, reason string, apply func([]FeedItem) ([]FeedItem, error)) []explainItemState {
	items := make([]FeedItem, 0, len(states))
	for _, state := range states {
		items = append(items, state.Item)
	}
	next, err := apply(items)
	if err != nil {
		for _, state := range states {
			appendExplainStage(records, state.RecordIndex, stageName, false, []string{stageName + ": explain simulation failed: " + err.Error()})
		}
		return nil
	}
	out := make([]explainItemState, 0, len(next))
	if len(next) == len(states) {
		for i, item := range next {
			state := states[i]
			state.Item = item
			appendExplainStage(records, state.RecordIndex, stageName, true, []string{reason})
			out = append(out, state)
		}
		return out
	}

	remaining := stateQueuesByKey(states)
	for _, item := range next {
		key := explainItemKey(item, -1)
		queue := remaining[key]
		if len(queue) == 0 {
			continue
		}
		state := queue[0]
		remaining[key] = queue[1:]
		state.Item = item
		appendExplainStage(records, state.RecordIndex, stageName, true, []string{reason})
		out = append(out, state)
	}
	return out
}

func explainMergeStage(states []explainItemState, records []ExplainRecord) []explainItemState {
	seen := map[string]struct{}{}
	out := make([]explainItemState, 0, len(states))
	for i, state := range states {
		key := explainItemKey(state.Item, i)
		if _, ok := seen[key]; ok {
			appendExplainStage(records, state.RecordIndex, "merge", false, []string{"merge: dropped duplicate key=" + key})
			continue
		}
		seen[key] = struct{}{}
		appendExplainStage(records, state.RecordIndex, "merge", true, []string{"merge: kept key=" + key})
		out = append(out, state)
	}
	return out
}

func explainRankStage(stage RankStage, states []explainItemState, records []ExplainRecord) []explainItemState {
	out := append([]explainItemState(nil), states...)
	rankBy := strings.ToLower(strings.TrimSpace(stage.By))
	switch rankBy {
	case "", "none":
	case "hotness":
		sortExplainStates(out, func(i, j FeedItem) bool {
			return parseMetadataScore(i, "hotness_score") > parseMetadataScore(j, "hotness_score")
		})
	case "published", "time":
		sortExplainStates(out, func(i, j FeedItem) bool {
			timeI, okI := i.PublishedTime()
			timeJ, okJ := j.PublishedTime()
			if !okI && !okJ {
				return false
			}
			if !okI {
				return false
			}
			if !okJ {
				return true
			}
			return timeI.After(timeJ)
		})
	default:
		for _, state := range states {
			appendExplainStage(records, state.RecordIndex, "rank", false, []string{"rank: unsupported rank " + stage.By})
		}
		return nil
	}
	for _, state := range out {
		appendExplainStage(records, state.RecordIndex, "rank", true, []string{"rank: by=" + rankBy})
	}
	return out
}

func sortExplainStates(states []explainItemState, less func(FeedItem, FeedItem) bool) {
	sort.SliceStable(states, func(i, j int) bool {
		return less(states[i].Item, states[j].Item)
	})
}

func explainDryRunStage(stageName string, states []explainItemState, records []ExplainRecord, reason string) []explainItemState {
	for _, state := range states {
		appendExplainStage(records, state.RecordIndex, stageName, true, []string{reason})
	}
	return states
}

func appendExplainStage(records []ExplainRecord, index int, stage string, kept bool, reasons []string) {
	records[index].Stages = append(records[index].Stages, ExplainStageRecord{
		Stage:   stage,
		Kept:    kept,
		Reasons: reasons,
	})
	if !kept {
		records[index].Kept = false
	}
	records[index].Reasons = append(records[index].Reasons, reasons...)
}

func flattenStageReasons(stages []ExplainStageRecord) []string {
	var reasons []string
	for _, stage := range stages {
		reasons = append(reasons, stage.Reasons...)
	}
	return reasons
}

func explainFilterRankItem(item FeedItem, cfg cliConfig, now time.Time) (bool, []string) {
	kept := true
	var reasons []string

	if len(cfg.includeKeyword) > 0 || len(cfg.excludeKeyword) > 0 || cfg.minKeywordScore > 0 {
		text := strings.ToLower(strings.Join([]string{item.Title, item.Description, item.Content, strings.Join(item.Categories, " ")}, " "))
		if matched := matchingContains(text, cfg.excludeKeyword); len(matched) > 0 {
			kept = false
			reasons = append(reasons, "keyword_filter: dropped by exclude="+strings.Join(matched, ","))
		}
		score := keywordScore(text, cfg.includeKeyword)
		if len(cfg.includeKeyword) > 0 && score == 0 {
			kept = false
			reasons = append(reasons, "keyword_filter: dropped by no include match")
		} else {
			reasons = append(reasons, fmt.Sprintf("keyword_filter: score=%d", score))
		}
		if cfg.minKeywordScore > 0 && score < cfg.minKeywordScore {
			kept = false
			reasons = append(reasons, fmt.Sprintf("keyword_filter: dropped by min_score=%d", cfg.minKeywordScore))
		}
	}

	if len(cfg.includeDomain) > 0 || len(cfg.excludeDomain) > 0 {
		domain := itemDomain(item)
		include := normalizeDomainList(cfg.includeDomain)
		exclude := normalizeDomainList(cfg.excludeDomain)
		switch {
		case domain == "" && len(include) > 0:
			kept = false
			reasons = append(reasons, "domain_filter: dropped by empty domain")
		case domainMatches(domain, exclude):
			kept = false
			reasons = append(reasons, "domain_filter: dropped by exclude domain="+domain)
		case len(include) > 0 && !domainMatches(domain, include):
			kept = false
			reasons = append(reasons, "domain_filter: dropped by include mismatch domain="+domain)
		default:
			reasons = append(reasons, "domain_filter: domain="+firstNonEmpty(domain, "(empty)"))
		}
	}

	if cfg.since > 0 {
		published, ok := item.PublishedTime()
		cutoff := now.Add(-cfg.since)
		if !ok {
			kept = false
			reasons = append(reasons, "time_window_filter: dropped by missing published time")
		} else if published.Before(cutoff) {
			kept = false
			reasons = append(reasons, "time_window_filter: dropped by older than "+cfg.since.String())
		} else {
			reasons = append(reasons, "time_window_filter: within "+cfg.since.String())
		}
	}

	rankBy := strings.ToLower(strings.TrimSpace(cfg.rankBy))
	if rankBy == "hotness" || cfg.minHotness > 0 {
		score := hotnessScore(item, now)
		reasons = append(reasons, "hotness_score: score="+formatScore(score))
		if cfg.minHotness > 0 && score < cfg.minHotness {
			kept = false
			reasons = append(reasons, fmt.Sprintf("min_hotness: dropped by min=%s", formatScore(cfg.minHotness)))
		}
	}
	if len(cfg.favUsers) > 0 {
		matched := matchingFavUsers(item, cfg.favUsers)
		if len(matched) == 0 {
			kept = false
			reasons = append(reasons, "fav_user: dropped by no matching user")
		} else {
			reasons = append(reasons, "fav_user: matched="+strings.Join(matched, ","))
		}
	}
	if rankBy != "" {
		reasons = append(reasons, "rank: by="+rankBy)
	}
	if len(reasons) == 0 {
		reasons = append(reasons, "no filter/rank criteria")
	}

	return kept, reasons
}

func explainKeywordStage(item FeedItem, stage KeywordFilterStage) (FeedItem, bool, []string) {
	kept := true
	var reasons []string
	text := strings.ToLower(strings.Join([]string{item.Title, item.Description, item.Content, strings.Join(item.Categories, " ")}, " "))
	if matched := matchingContains(text, stage.Exclude); len(matched) > 0 {
		kept = false
		reasons = append(reasons, "keyword_filter: dropped by exclude="+strings.Join(matched, ","))
	}
	score := keywordScore(text, stage.Include)
	if len(stage.Include) > 0 && score == 0 {
		kept = false
		reasons = append(reasons, "keyword_filter: dropped by no include match")
	} else {
		reasons = append(reasons, fmt.Sprintf("keyword_filter: score=%d", score))
	}
	if stage.MinScore > 0 && score < stage.MinScore {
		kept = false
		reasons = append(reasons, fmt.Sprintf("keyword_filter: dropped by min_score=%d", stage.MinScore))
	}
	item.Metadata = cloneMetadata(item.Metadata)
	item.Metadata["keyword_score"] = strconv.Itoa(score)
	return item, kept, reasons
}

func explainDomainStage(item FeedItem, stage DomainFilterStage) (FeedItem, bool, []string) {
	kept := true
	domain := itemDomain(item)
	include := normalizeDomainList(stage.Include)
	exclude := normalizeDomainList(stage.Exclude)
	var reasons []string
	switch {
	case domain == "" && len(include) > 0:
		kept = false
		reasons = append(reasons, "domain_filter: dropped by empty domain")
	case domainMatches(domain, exclude):
		kept = false
		reasons = append(reasons, "domain_filter: dropped by exclude domain="+domain)
	case len(include) > 0 && !domainMatches(domain, include):
		kept = false
		reasons = append(reasons, "domain_filter: dropped by include mismatch domain="+domain)
	default:
		reasons = append(reasons, "domain_filter: domain="+firstNonEmpty(domain, "(empty)"))
	}
	if domain != "" {
		item.Metadata = cloneMetadata(item.Metadata)
		item.Metadata["domain"] = domain
	}
	return item, kept, reasons
}

func explainTimeWindowStage(item FeedItem, stage TimeWindowFilterStage) (FeedItem, bool, []string) {
	if stage.Since <= 0 {
		return item, true, []string{"time_window_filter: no window"}
	}
	now := stage.Now
	if now.IsZero() {
		now = time.Now()
	}
	published, ok := item.PublishedTime()
	cutoff := now.Add(-stage.Since)
	if !ok {
		return item, false, []string{"time_window_filter: dropped by missing published time"}
	}
	if published.Before(cutoff) {
		return item, false, []string{"time_window_filter: dropped by older than " + stage.Since.String()}
	}
	return item, true, []string{"time_window_filter: within " + stage.Since.String()}
}

func explainMinHotnessStage(item FeedItem, stage MinHotnessStage) (FeedItem, bool, []string) {
	score := parseMetadataScore(item, "hotness_score")
	if stage.Min > 0 && score < stage.Min {
		return item, false, []string{fmt.Sprintf("min_hotness: dropped by min=%s score=%s", formatScore(stage.Min), formatScore(score))}
	}
	return item, true, []string{"min_hotness: score=" + formatScore(score)}
}

func explainFavUserStage(item FeedItem, stage FavUserStage) (FeedItem, bool, []string) {
	matched := matchingFavUsers(item, stage.Users)
	if len(stage.Users) > 0 && len(matched) == 0 {
		return item, false, []string{"fav_user: dropped by no matching user"}
	}
	if len(matched) > 0 {
		item.Metadata = cloneMetadata(item.Metadata)
		item.Metadata["fav_users"] = strings.Join(matched, ",")
		return item, true, []string{"fav_user: matched=" + strings.Join(matched, ",")}
	}
	return item, true, []string{"fav_user: no users configured"}
}

func stateQueuesByKey(states []explainItemState) map[string][]explainItemState {
	out := make(map[string][]explainItemState, len(states))
	for i, state := range states {
		key := explainItemKey(state.Item, i)
		out[key] = append(out[key], state)
	}
	return out
}

func explainItemKey(item FeedItem, index int) string {
	key := feedItemDedupKey(item)
	if key == "" && index >= 0 {
		return fmt.Sprintf("__index_%d", index)
	}
	return key
}

func matchingContains(text string, patterns []string) []string {
	var matched []string
	for _, pattern := range patterns {
		pattern = strings.ToLower(strings.TrimSpace(pattern))
		if pattern != "" && strings.Contains(text, pattern) {
			matched = append(matched, pattern)
		}
	}
	return matched
}

func matchingFavUsers(item FeedItem, users []string) []string {
	watch := make(map[string]struct{}, len(users))
	for _, user := range users {
		user = strings.ToLower(strings.TrimSpace(user))
		if user != "" {
			watch[user] = struct{}{}
		}
	}
	var matched []string
	for _, comment := range item.HatenaBookmarkComments {
		if _, ok := watch[strings.ToLower(comment.User)]; ok {
			matched = append(matched, comment.User)
		}
	}
	return uniqueStrings(matched)
}

func stageNames(stages []FeedItemStage) []string {
	names := make([]string, 0, len(stages))
	for _, stage := range stages {
		names = append(names, stage.Name())
	}
	return names
}

func explainOutputName(cfg cliConfig, cmd string) string {
	output := operationName(cfg.operation, cmd)
	if output == "" || output == commandName([]string{cmd}) || output == "feed2cli" {
		return "rss"
	}
	return output
}
