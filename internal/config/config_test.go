package config

import (
	"path/filepath"
	"testing"
	"time"
)

func TestBuildRuleInference(test *testing.T) {
	test.Run("no flags -> block", func(test *testing.T) {
		rule, err := BuildRule(RuleSpec{Domain: "a.com"})
		if err != nil || rule.Kind != RuleTypeBlock {
			test.Fatalf("got kind=%q err=%v", rule.Kind, err)
		}
	})

	test.Run("window -> scheduled", func(test *testing.T) {
		rule, err := BuildRule(RuleSpec{Domain: "a.com", Windows: []TimeWindow{{"09:00", "17:00"}}})
		if err != nil || rule.Kind != RuleTypeScheduled || len(rule.BlockedWindows) != 1 {
			test.Fatalf("got kind=%q windows=%v err=%v", rule.Kind, rule.BlockedWindows, err)
		}
	})

	test.Run("for -> timer", func(test *testing.T) {
		rule, err := BuildRule(RuleSpec{Domain: "a.com", For: 90 * time.Minute})
		if err != nil || rule.Kind != RuleTypeTimer || !rule.BlockedUntil.After(time.Now()) {
			test.Fatalf("got kind=%q until=%v err=%v", rule.Kind, rule.BlockedUntil, err)
		}
	})

	test.Run("until in past rolls to next day", func(test *testing.T) {
		past := time.Now().Add(-2 * time.Hour).Format("15:04")
		rule, err := BuildRule(RuleSpec{Domain: "a.com", Until: past})
		if err != nil || !rule.BlockedUntil.After(time.Now()) {
			test.Fatalf("expected future, got %v err=%v", rule.BlockedUntil, err)
		}
	})

	test.Run("multiple type flags -> error", func(test *testing.T) {
		if _, err := BuildRule(RuleSpec{Domain: "a.com", For: time.Hour, Until: "10:00"}); err == nil {
			test.Fatal("expected error for conflicting type flags")
		}
	})

	test.Run("bad window -> error", func(test *testing.T) {
		if _, err := BuildRule(RuleSpec{Domain: "a.com", Windows: []TimeWindow{{"9am", "5pm"}}}); err == nil {
			test.Fatal("expected error for malformed window")
		}
	})
}

func TestAppendRuleOutcomes(test *testing.T) {
	DomainsConfigPath = filepath.Join(test.TempDir(), "domainconfig.toml")
	if err := writeConfig(DomainsConfigPath, nil); err != nil {
		test.Fatal(err)
	}

	sched := func(ws ...TimeWindow) DomainRule {
		return DomainRule{Domain: "x.com", Kind: RuleTypeScheduled, BlockedWindows: ws}
	}

	// new domain
	res, err := AppendRule(sched(TimeWindow{"09:00", "17:00"}), false)
	if err != nil || res.Outcome != OutcomeAdded {
		test.Fatalf("add: outcome=%d err=%v", res.Outcome, err)
	}

	// scheduled + scheduled merges new window
	res, err = AppendRule(sched(TimeWindow{"20:00", "22:00"}), false)
	if err != nil || res.Outcome != OutcomeWindowsMerged || res.AddedWindows != 1 {
		test.Fatalf("merge: outcome=%d added=%d err=%v", res.Outcome, res.AddedWindows, err)
	}

	// duplicate window is a no-op merge
	res, _ = AppendRule(sched(TimeWindow{"20:00", "22:00"}), false)
	if res.Outcome != OutcomeWindowsMerged || res.AddedWindows != 0 {
		test.Fatalf("dup merge: outcome=%d added=%d", res.Outcome, res.AddedWindows)
	}

	// block over existing scheduled -> conflict, nothing written
	res, _ = AppendRule(DomainRule{Domain: "x.com", Kind: RuleTypeBlock}, false)
	if res.Outcome != OutcomeConflict {
		test.Fatalf("conflict expected, got %d", res.Outcome)
	}
	rules, _ := LoadConfig(DomainsConfigPath)
	if rules[0].Kind != RuleTypeScheduled {
		test.Fatalf("conflict should not mutate, kind=%q", rules[0].Kind)
	}

	// replace overrides regardless of kind, preserves AddedAt
	added := rules[0].AddedAt
	res, _ = AppendRule(DomainRule{Domain: "x.com", Kind: RuleTypeBlock}, true)
	if res.Outcome != OutcomeReplaced {
		test.Fatalf("replace expected, got %d", res.Outcome)
	}
	rules, _ = LoadConfig(DomainsConfigPath)
	if rules[0].Kind != RuleTypeBlock || !rules[0].AddedAt.Equal(added) {
		test.Fatalf("replace: kind=%q addedAt preserved=%v", rules[0].Kind, rules[0].AddedAt.Equal(added))
	}
}

func TestSealStateSurvivesRuleEdits(test *testing.T) {
	DomainsConfigPath = filepath.Join(test.TempDir(), "domainconfig.toml")
	if err := writeConfig(DomainsConfigPath, nil); err != nil {
		test.Fatal(err)
	}

	// Fresh config is unsealed.
	if sealed, err := IsSealed(); err != nil || sealed {
		test.Fatalf("fresh config: sealed=%v err=%v", sealed, err)
	}

	if err := SetSealed(true); err != nil {
		test.Fatal(err)
	}
	if sealed, _ := IsSealed(); !sealed {
		test.Fatal("expected sealed after SetSealed(true)")
	}

	// A rule edit must not clobber the seal.
	if _, err := AppendRule(DomainRule{Domain: "a.com", Kind: RuleTypeBlock}, false); err != nil {
		test.Fatal(err)
	}
	if sealed, _ := IsSealed(); !sealed {
		test.Fatal("AppendRule must preserve seal")
	}
	if _, err := ReleaseRule("a.com"); err != nil {
		test.Fatal(err)
	}
	if sealed, _ := IsSealed(); !sealed {
		test.Fatal("ReleaseRule must preserve seal")
	}

	// SetSealed preserves existing rules.
	if _, err := AppendRule(DomainRule{Domain: "b.com", Kind: RuleTypeBlock}, false); err != nil {
		test.Fatal(err)
	}
	if err := SetSealed(false); err != nil {
		test.Fatal(err)
	}
	if sealed, _ := IsSealed(); sealed {
		test.Fatal("expected unsealed after SetSealed(false)")
	}
	rules, _ := LoadConfig(DomainsConfigPath)
	if len(rules) != 1 || rules[0].Domain != "b.com" {
		test.Fatalf("SetSealed must preserve rules, got %+v", rules)
	}
}
