package ec2

import "testing"

func TestApplySafetyDefaultsCapsDiskFill(t *testing.T) {
	config := Configuration{
		EC2Config: EC2Config{
			MaxDiskFillMB: 2048,
		},
	}

	got := ApplySafetyDefaults(config, 256, false)
	if got.EC2Config.MaxDiskFillMB != 256 {
		t.Fatalf("MaxDiskFillMB = %d, want 256", got.EC2Config.MaxDiskFillMB)
	}
}

func TestApplySafetyDefaultsPreservesLowerDiskFillLimit(t *testing.T) {
	config := Configuration{
		EC2Config: EC2Config{
			MaxDiskFillMB: 128,
		},
	}

	got := ApplySafetyDefaults(config, 256, true)
	if got.EC2Config.MaxDiskFillMB != 128 {
		t.Fatalf("MaxDiskFillMB = %d, want 128", got.EC2Config.MaxDiskFillMB)
	}
	if !got.EC2Config.KeepDiskFile {
		t.Fatal("KeepDiskFile = false, want true")
	}
}
