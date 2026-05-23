package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

type PipelineConfig struct {
	Input  PipelineInputConfig  `json:"input,omitempty"`
	Output PipelineOutputConfig `json:"output,omitempty"`
	Stages []PipelineStageSpec  `json:"stages,omitempty"`
}

type PipelineInputConfig struct {
	Format string   `json:"format,omitempty"`
	URLs   []string `json:"urls,omitempty"`
}

type PipelineOutputConfig struct {
	Type                string `json:"type,omitempty"`
	DigestTitle         string `json:"digest_title,omitempty"`
	DigestWindow        string `json:"digest_window,omitempty"`
	StateBackend        string `json:"state_backend,omitempty"`
	StatePath           string `json:"state_path,omitempty"`
	SlackChannel        string `json:"slack_channel,omitempty"`
	SlackDryRun         bool   `json:"slack_dry_run,omitempty"`
	SlackSkipValidation bool   `json:"slack_skip_validate,omitempty"`
}

type PipelineStageSpec struct {
	Type          string   `json:"type"`
	Name          string   `json:"name,omitempty"`
	Include       []string `json:"include,omitempty"`
	Exclude       []string `json:"exclude,omitempty"`
	MinScore      int      `json:"min_score,omitempty"`
	Since         string   `json:"since,omitempty"`
	Users         []string `json:"users,omitempty"`
	By            string   `json:"by,omitempty"`
	Min           float64  `json:"min,omitempty"`
	DefaultSource string   `json:"default_source,omitempty"`
	Command       string   `json:"command,omitempty"`
	Args          []string `json:"args,omitempty"`
	Env           []string `json:"env,omitempty"`
	Timeout       string   `json:"timeout,omitempty"`
	StderrLimit   int      `json:"stderr_limit,omitempty"`
}

func LoadPipelineConfig(path string) (PipelineConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return PipelineConfig{}, err
	}
	var cfg PipelineConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return PipelineConfig{}, fmt.Errorf("pipeline config must be JSON for now: %w", err)
	}
	return cfg, nil
}

func applyPipelineConfig(cfg cliConfig, pipeline PipelineConfig) (cliConfig, error) {
	if input := strings.ToLower(strings.TrimSpace(pipeline.Input.Format)); input != "" && cfg.inputFormat == "feed" {
		cfg.inputFormat = input
	}
	cfg.feedURLs = append(pipeline.Input.URLs, cfg.feedURLs...)

	if output := strings.ToLower(strings.TrimSpace(pipeline.Output.Type)); output != "" && cfg.operation == "" {
		cfg.operation = output
	}
	if pipeline.Output.DigestTitle != "" && cfg.digestTitle == "feed2cli digest" {
		cfg.digestTitle = pipeline.Output.DigestTitle
	}
	if pipeline.Output.DigestWindow != "" && cfg.digestWindow == 24*time.Hour {
		window, err := time.ParseDuration(pipeline.Output.DigestWindow)
		if err != nil {
			return cliConfig{}, fmt.Errorf("invalid output.digest_window %q: %w", pipeline.Output.DigestWindow, err)
		}
		cfg.digestWindow = window
	}
	if pipeline.Output.StateBackend != "" && cfg.stateBackend == envDefault("FEED2CLI_STATE_BACKEND", "json") {
		cfg.stateBackend = strings.ToLower(strings.TrimSpace(pipeline.Output.StateBackend))
	}
	if pipeline.Output.StatePath != "" && cfg.statePath == defaultStatePath(cfg.stateBackend) {
		cfg.statePath = pipeline.Output.StatePath
	}
	if pipeline.Output.SlackChannel != "" && cfg.slackChannel == "" {
		cfg.slackChannel = pipeline.Output.SlackChannel
	}
	if pipeline.Output.SlackDryRun {
		cfg.slackDryRun = true
	}
	if pipeline.Output.SlackSkipValidation {
		cfg.slackSkipValidate = true
	}

	stages, err := PipelineStagesFromConfig(pipeline)
	if err != nil {
		return cliConfig{}, err
	}
	cfg.pipelineStages = stages
	return cfg, nil
}

func defaultStatePath(backend string) string {
	if strings.ToLower(strings.TrimSpace(backend)) == "sqlite" {
		return "hatena_state.sqlite"
	}
	return stateFilePath
}

func PipelineStagesFromConfig(config PipelineConfig) ([]FeedItemStage, error) {
	stages := make([]FeedItemStage, 0, len(config.Stages))
	for _, spec := range config.Stages {
		stage, err := pipelineStageFromSpec(spec)
		if err != nil {
			return nil, err
		}
		stages = append(stages, stage)
	}
	return stages, nil
}

func pipelineStageFromSpec(spec PipelineStageSpec) (FeedItemStage, error) {
	stageType := strings.ToLower(strings.TrimSpace(firstNonEmpty(spec.Type, spec.Name)))
	switch stageType {
	case "normalize":
		return NormalizeStage{}, nil
	case "merge":
		return MergeStage{}, nil
	case "hatena_bookmark", "hatena-bookmark", "hatena":
		return HatenaBookmarkStage{}, nil
	case "keyword_filter", "keyword-filter", "keyword":
		return KeywordFilterStage{Include: spec.Include, Exclude: spec.Exclude, MinScore: spec.MinScore}, nil
	case "domain_filter", "domain-filter", "domain":
		return DomainFilterStage{Include: spec.Include, Exclude: spec.Exclude}, nil
	case "time_window", "time-window", "since":
		since, err := parsePipelineDuration(spec.Since, stageType)
		if err != nil {
			return nil, err
		}
		return TimeWindowFilterStage{Since: since}, nil
	case "hotness_score", "hotness-score", "hotness":
		return HotnessScoreStage{}, nil
	case "min_hotness", "min-hotness":
		return MinHotnessStage{Min: spec.Min}, nil
	case "fav_user", "fav-user":
		return FavUserStage{Users: spec.Users}, nil
	case "rank":
		return RankStage{By: spec.By}, nil
	case "source_label", "source-label", "source":
		return SourceLabelStage{DefaultSource: spec.DefaultSource}, nil
	case "tag", "tags":
		return TagEnrichStage{}, nil
	case "ogp":
		return OGPEnrichStage{}, nil
	case "content", "readability":
		return ContentFetchStage{}, nil
	case "summary", "local_summary", "local-llm-summary":
		return SummaryStage{}, nil
	case "plugin", "external", "external_command", "external-command", "command":
		timeout, err := parseOptionalPipelineDuration(spec.Timeout)
		if err != nil {
			return nil, fmt.Errorf("invalid timeout for stage %q: %w", firstNonEmpty(spec.Type, spec.Name), err)
		}
		return ExternalCommandStage{
			StageName:   firstNonEmpty(spec.Name, spec.Type),
			Command:     spec.Command,
			Args:        append([]string(nil), spec.Args...),
			Env:         append([]string(nil), spec.Env...),
			Timeout:     timeout,
			StderrLimit: spec.StderrLimit,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported pipeline stage %q", firstNonEmpty(spec.Type, spec.Name))
	}
}

func parsePipelineDuration(value, field string) (time.Duration, error) {
	if strings.TrimSpace(value) == "" {
		return 0, fmt.Errorf("%s stage requires since", field)
	}
	duration, err := time.ParseDuration(value)
	if err != nil {
		return 0, err
	}
	return duration, nil
}

func parseOptionalPipelineDuration(value string) (time.Duration, error) {
	if strings.TrimSpace(value) == "" {
		return 0, nil
	}
	return time.ParseDuration(value)
}
