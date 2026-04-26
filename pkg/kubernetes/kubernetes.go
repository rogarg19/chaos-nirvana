package kubernetes

import (
	"bytes"
	"context"
	"fmt"
	"math"
	"os/exec"
	"strings"

	"github.com/rogarg19/chaos-nirvana/pkg/scenario"
)

type Chaos struct {
	kubectl string
}

func New() *Chaos {
	return &Chaos{kubectl: "kubectl"}
}

func (c *Chaos) Run(ctx context.Context, s scenario.Scenario) error {
	if s.Action != scenario.ActionKubernetesKillPods {
		return fmt.Errorf("unsupported kubernetes action %q", s.Action)
	}
	if s.Target.Namespace == "" {
		s.Target.Namespace = "default"
	}
	if s.Target.Selector == "" {
		return fmt.Errorf("selector is required to select pods")
	}

	pods, err := c.listPods(ctx, s.Target.Namespace, s.Target.Selector)
	if err != nil {
		return err
	}
	if len(pods) == 0 {
		return fmt.Errorf("no pods matched selector %q in namespace %q", s.Target.Selector, s.Target.Namespace)
	}

	count := s.Parameters.Count
	if count <= 0 {
		percent := s.Parameters.Percent
		if percent <= 0 {
			percent = 100
		}
		count = int(math.Ceil(float64(len(pods)) * float64(percent) / 100.0))
		if count < 1 {
			count = 1
		}
	}
	if count > len(pods) {
		count = len(pods)
	}

	for _, pod := range pods[:count] {
		if err := c.deletePod(ctx, s.Target.Namespace, pod); err != nil {
			return err
		}
	}
	return nil
}

func (c *Chaos) listPods(ctx context.Context, namespace, selector string) ([]string, error) {
	args := []string{"-n", namespace, "get", "pods", "-l", selector, "-o", "name"}
	output, err := c.run(ctx, args...)
	if err != nil {
		return nil, fmt.Errorf("list pods: %w", err)
	}
	lines := strings.Split(strings.TrimSpace(output), "\n")
	pods := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		pods = append(pods, strings.TrimPrefix(line, "pod/"))
	}
	return pods, nil
}

func (c *Chaos) deletePod(ctx context.Context, namespace, pod string) error {
	_, err := c.run(ctx, "-n", namespace, "delete", "pod", pod)
	if err != nil {
		return fmt.Errorf("delete pod %q: %w", pod, err)
	}
	return nil
}

func (c *Chaos) run(ctx context.Context, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, c.kubectl, args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	output, err := cmd.Output()
	if err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg != "" {
			return "", fmt.Errorf("%v: %s", err, msg)
		}
		return "", err
	}
	return string(output), nil
}
