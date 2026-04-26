package nlp

import (
	"testing"
	"time"

	"github.com/rogarg19/chaos-nirvana/pkg/scenario"
)

func TestParseKubernetesKillPods(t *testing.T) {
	parsed, err := Parse("Kill 50% pods of service payments in namespace checkout")
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	if parsed.Service != scenario.ServiceKubernetes {
		t.Fatalf("service = %q, want %q", parsed.Service, scenario.ServiceKubernetes)
	}
	if parsed.Action != scenario.ActionKubernetesKillPods {
		t.Fatalf("action = %q, want %q", parsed.Action, scenario.ActionKubernetesKillPods)
	}
	if parsed.Parameters.Percent != 50 {
		t.Fatalf("percent = %d, want 50", parsed.Parameters.Percent)
	}
	if parsed.Target.Namespace != "checkout" {
		t.Fatalf("namespace = %q, want checkout", parsed.Target.Namespace)
	}
	if parsed.Target.Selector != "app=payments" {
		t.Fatalf("selector = %q, want app=payments", parsed.Target.Selector)
	}
}

func TestParseRedisCPUSpike(t *testing.T) {
	parsed, err := Parse("put load on redis cluster such that CPU usage spikes to 90% for more than 15 minutes")
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	if parsed.Service != scenario.ServiceRedis {
		t.Fatalf("service = %q, want %q", parsed.Service, scenario.ServiceRedis)
	}
	if parsed.Action != scenario.ActionRedisCPUSpike {
		t.Fatalf("action = %q, want %q", parsed.Action, scenario.ActionRedisCPUSpike)
	}
	if !parsed.Parameters.Cluster {
		t.Fatal("cluster = false, want true")
	}
	if parsed.Parameters.CPUPercent != 90 {
		t.Fatalf("cpu percent = %d, want 90", parsed.Parameters.CPUPercent)
	}
	if parsed.Duration != 15*time.Minute {
		t.Fatalf("duration = %s, want 15m", parsed.Duration)
	}
}
