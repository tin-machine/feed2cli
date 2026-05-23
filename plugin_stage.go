package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

type ExternalCommandStage struct {
	StageName   string
	Command     string
	Args        []string
	Env         []string
	Timeout     time.Duration
	StderrLimit int
}

func (s ExternalCommandStage) Name() string {
	if strings.TrimSpace(s.StageName) != "" {
		return s.StageName
	}
	return "external_command"
}

func (s ExternalCommandStage) Apply(ctx context.Context, items []FeedItem) ([]FeedItem, error) {
	command := strings.TrimSpace(s.Command)
	if command == "" {
		return nil, errors.New("external command stage requires command")
	}

	runCtx := ctx
	cancel := func() {}
	if s.Timeout > 0 {
		runCtx, cancel = context.WithTimeout(ctx, s.Timeout)
	}
	defer cancel()

	var stdin bytes.Buffer
	if err := OutputJSONLTo(&stdin, items); err != nil {
		return nil, fmt.Errorf("%s stdin encode failed: %w", s.Name(), err)
	}

	cmd := exec.CommandContext(runCtx, command, s.Args...)
	cmd.Stdin = &stdin
	if len(s.Env) > 0 {
		cmd.Env = append(os.Environ(), s.Env...)
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if runCtx.Err() != nil {
		return nil, fmt.Errorf("%s timed out: %w", s.Name(), runCtx.Err())
	}
	if err != nil {
		return nil, fmt.Errorf("%s failed: %w%s", s.Name(), err, stderrSuffix(stderr.String(), s.StderrLimit))
	}

	out, err := ReadJSONLItems(&stdout)
	if err != nil {
		return nil, fmt.Errorf("%s stdout decode failed: %w%s", s.Name(), err, stderrSuffix(stderr.String(), s.StderrLimit))
	}
	return out, nil
}

func stderrSuffix(stderr string, limit int) string {
	stderr = strings.TrimSpace(stderr)
	if stderr == "" {
		return ""
	}
	if limit <= 0 {
		limit = 4096
	}
	runes := []rune(stderr)
	if len(runes) > limit {
		stderr = string(runes[:limit]) + "..."
	}
	return ": stderr: " + stderr
}
