package kubernetes

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/rogarg19/chaos-nirvana/pkg/scenario"
)

func TestPlanRejectsPodCountAboveMax(t *testing.T) {
	kubectl := fakeKubectl(t, "pod/payments-1\npod/payments-2\npod/payments-3\n")
	chaos := &Chaos{kubectl: kubectl}

	_, err := chaos.Plan(context.Background(), scenario.Scenario{
		Action: scenario.ActionKubernetesKillPods,
		Target: scenario.Target{
			Namespace: "checkout",
			Selector:  "app=payments",
		},
		Parameters: scenario.Parameters{Percent: 100},
	}, 2)
	if err == nil {
		t.Fatal("Plan returned nil error, want max pod limit error")
	}
}

func TestPlanSelectsRequestedPodCount(t *testing.T) {
	kubectl := fakeKubectl(t, "pod/payments-1\npod/payments-2\npod/payments-3\n")
	chaos := &Chaos{kubectl: kubectl}

	plan, err := chaos.Plan(context.Background(), scenario.Scenario{
		Action: scenario.ActionKubernetesKillPods,
		Target: scenario.Target{
			Namespace: "checkout",
			Selector:  "app=payments",
		},
		Parameters: scenario.Parameters{Count: 2},
	}, 5)
	if err != nil {
		t.Fatalf("Plan returned error: %v", err)
	}
	if len(plan.MatchedPods) != 3 {
		t.Fatalf("matched pods = %d, want 3", len(plan.MatchedPods))
	}
	if len(plan.TargetPods) != 2 {
		t.Fatalf("target pods = %d, want 2", len(plan.TargetPods))
	}
}

func fakeKubectl(t *testing.T, output string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "kubectl")
	script := "#!/bin/sh\nprintf '%s' '" + output + "'\n"
	if err := os.WriteFile(path, []byte(script), 0755); err != nil {
		t.Fatalf("write fake kubectl: %v", err)
	}
	return path
}
