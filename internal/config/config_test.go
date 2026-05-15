package config

import (
	"path/filepath"
	"testing"
	"time"
)

func TestBuildRuleInference(t *testing.T) {
	t.Run("no flags -> block", func(t *testing.T) {
		r, err := BuildRule(RuleSpec{Domain: "a.com"})
		if err != nil || r.Kind != RuleTypeBlock {
			t.Fatalf("got kind=%q err=%v", r.Kind, err)
		}
	})

	t.Run("window -> scheduled", func(t *testing.T) {
		r, err := BuildRule(RuleSpec{Domain: "a.com", Windows: []TimeWindow{{"09:00", "17:00"}}})
		if err != nil || r.Kind != RuleTypeScheduled || len(r.BlockedWindows) != 1 {
			t.Fatalf("got kind=%q windows=%v err=%v", r.Kind, r.BlockedWindows, err)
		}
	})

	t.Run("for -> timer", func(t *testing.T) {
		r, err := BuildRule(RuleSpec{Domain: "a.com", For: 90 * time.Minute})
		if err != nil || r.Kind != RuleTypeTimer || !r.BlockedUntil.After(time.Now()) {
			t.Fatalf("got kind=%q until=%v err=%v", r.Kind, r.BlockedUntil, err)
		}
	})

	t.Run("until in past rolls to next day", func(t *testing.T) {
		past := time.Now().Add(-2 * time.Hour).Format("15:04")
		r, err := BuildRule(RuleSpec{Domain: "a.com", Until: past})
		if err != nil || !r.BlockedUntil.After(time.Now()) {
			t.Fatalf("expected future, got %v err=%v", r.BlockedUntil, err)
		}
	})

	t.Run("multiple type flags -> error", func(t *testing.T) {
		if _, err := BuildRule(RuleSpec{Domain: "a.com", For: time.Hour, Until: "10:00"}); err == nil {
			t.Fatal("expected error for conflicting type flags")
		}
	})

	t.Run("bad window -> error", func(t *testing.T) {
		if _, err := BuildRule(RuleSpec{Domain: "a.com", Windows: []TimeWindow{{"9am", "5pm"}}}); err == nil {
			t.Fatal("expected error for malformed window")
		}
	})
}

func TestAppendRuleOutcomes(t *testing.T) {
	DomainsConfigPath = filepath.Join(t.TempDir(), "domainconfig.toml")
	if err := writeConfig(DomainsConfigPath, nil); err != nil {
		t.Fatal(err)
	}

	sched := func(ws ...TimeWindow) DomainRule {
		return DomainRule{Domain: "x.com", Kind: RuleTypeScheduled, BlockedWindows: ws}
	}

	// new domain
	res, err := AppendRule(sched(TimeWindow{"09:00", "17:00"}), false)
	if err != nil || res.Outcome != OutcomeAdded {
		t.Fatalf("add: outcome=%d err=%v", res.Outcome, err)
	}

	// scheduled + scheduled merges new window
	res, err = AppendRule(sched(TimeWindow{"20:00", "22:00"}), false)
	if err != nil || res.Outcome != OutcomeWindowsMerged || res.AddedWindows != 1 {
		t.Fatalf("merge: outcome=%d added=%d err=%v", res.Outcome, res.AddedWindows, err)
	}

	// duplicate window is a no-op merge
	res, _ = AppendRule(sched(TimeWindow{"20:00", "22:00"}), false)
	if res.Outcome != OutcomeWindowsMerged || res.AddedWindows != 0 {
		t.Fatalf("dup merge: outcome=%d added=%d", res.Outcome, res.AddedWindows)
	}

	// block over existing scheduled -> conflict, nothing written
	res, _ = AppendRule(DomainRule{Domain: "x.com", Kind: RuleTypeBlock}, false)
	if res.Outcome != OutcomeConflict {
		t.Fatalf("conflict expected, got %d", res.Outcome)
	}
	rules, _ := LoadConfig(DomainsConfigPath)
	if rules[0].Kind != RuleTypeScheduled {
		t.Fatalf("conflict should not mutate, kind=%q", rules[0].Kind)
	}

	// replace overrides regardless of kind, preserves AddedAt
	added := rules[0].AddedAt
	res, _ = AppendRule(DomainRule{Domain: "x.com", Kind: RuleTypeBlock}, true)
	if res.Outcome != OutcomeReplaced {
		t.Fatalf("replace expected, got %d", res.Outcome)
	}
	rules, _ = LoadConfig(DomainsConfigPath)
	if rules[0].Kind != RuleTypeBlock || !rules[0].AddedAt.Equal(added) {
		t.Fatalf("replace: kind=%q addedAt preserved=%v", rules[0].Kind, rules[0].AddedAt.Equal(added))
	}
}
