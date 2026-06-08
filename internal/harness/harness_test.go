package harness

import "testing"

func TestHarnessDeclaresProjectGoalAndRuntimeRule(t *testing.T) {
	if ProjectName != "ExitFleet" {
		t.Fatalf("ProjectName = %q, want ExitFleet", ProjectName)
	}
	if ProjectGoal == "" {
		t.Fatal("ProjectGoal must not be empty")
	}
	if RuntimeRule != "One active exit node maps to exactly one worker container." {
		t.Fatalf("RuntimeRule = %q", RuntimeRule)
	}
}
