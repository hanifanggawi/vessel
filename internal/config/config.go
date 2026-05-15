package config

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/hanifanggawi/vessel/internal/hosts"
)

var (
	DomainsConfigPath string
	EtcHostsPath      string
)

func Init() error {
	cwd, _ := os.Getwd()
	var err error

	if path, ok := os.LookupEnv("VESSEL_DOMAINS_CONFIG_PATH"); ok {
		DomainsConfigPath = path
	} else {
		DomainsConfigPath = filepath.Join(cwd, ".local", "domainconfig.toml")
	}

	if path, ok := os.LookupEnv("VESSEL_HOSTS_PATH"); ok {
		EtcHostsPath = path
	} else {
		EtcHostsPath, err = hosts.GetHostsPath()
		if err != nil {
			return err
		}
	}
	return nil
}

type RuleType string

const (
	RuleTypeBlock     RuleType = "block"
	RuleTypeScheduled RuleType = "scheduled"
	RuleTypeTimer     RuleType = "timer"
)

type TimeWindow struct {
	StartTime string `toml:"start"`
	EndTime   string `toml:"end"`
}

func parseTimeOfDay(s string) (time.Time, error) {
	t, err := time.Parse("15:04", s)
	if err != nil {
		return time.Time{}, err
	}
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), t.Hour(), t.Minute(), 0, 0, now.Location()), nil
}

func (w TimeWindow) equals(other TimeWindow) bool {
	return w.StartTime == other.StartTime && w.EndTime == other.EndTime
}

func (w TimeWindow) ParseStart() (time.Time, error) { return parseTimeOfDay(w.StartTime) }
func (w TimeWindow) ParseEnd() (time.Time, error)   { return parseTimeOfDay(w.EndTime) }

type DomainRule struct {
	Domain         string
	AddedAt        time.Time    `toml:"added_at"`
	Kind           RuleType     `toml:"kind"`
	BlockedUntil   time.Time    `toml:"blocked_until"`
	BlockedWindows []TimeWindow `toml:"blocked_windows"`
}

func (r DomainRule) ConfigStr() string {
	parts := []string{fmt.Sprintf("kind = %q", r.Kind)}

	if !r.AddedAt.IsZero() {
		parts = append(parts, fmt.Sprintf("added_at = %q", r.AddedAt.Format(time.RFC3339)))
	}

	if !r.BlockedUntil.IsZero() {
		parts = append(parts, fmt.Sprintf("blocked_until = %q", r.BlockedUntil.Format(time.RFC3339)))
	}

	if len(r.BlockedWindows) > 0 {
		windows := make([]string, len(r.BlockedWindows))
		for i, w := range r.BlockedWindows {
			windows[i] = fmt.Sprintf("{start=%q, end=%q}", w.StartTime, w.EndTime)
		}
		parts = append(parts, fmt.Sprintf("blocked_windows = [%s]", strings.Join(windows, ", ")))
	}

	return fmt.Sprintf("%q = { %s }", r.Domain, strings.Join(parts, ", "))
}

// RuleSpec is the raw, flag-level intent for a rule. The concrete RuleType is
// inferred from which fields are set: windows -> scheduled, for/until -> timer,
// none -> block.
type RuleSpec struct {
	Domain  string
	Windows []TimeWindow
	For     time.Duration
	Until   string
}

func (s RuleSpec) typeCount() int {
	n := 0
	if len(s.Windows) > 0 {
		n++
	}
	if s.For > 0 {
		n++
	}
	if s.Until != "" {
		n++
	}
	return n
}

// BuildRule infers the RuleType from the spec and produces a concrete
// DomainRule, validating the time inputs.
func BuildRule(spec RuleSpec) (DomainRule, error) {
	if spec.Domain == "" {
		return DomainRule{}, fmt.Errorf("domain is required")
	}
	if spec.typeCount() > 1 {
		return DomainRule{}, fmt.Errorf("a rule can only be one of: scheduled (--window), timer (--for / --until)")
	}

	rule := DomainRule{Domain: spec.Domain, AddedAt: time.Now()}

	switch {
	case len(spec.Windows) > 0:
		for _, w := range spec.Windows {
			if _, err := w.ParseStart(); err != nil {
				return DomainRule{}, fmt.Errorf("invalid window start %q: expected HH:MM", w.StartTime)
			}
			if _, err := w.ParseEnd(); err != nil {
				return DomainRule{}, fmt.Errorf("invalid window end %q: expected HH:MM", w.EndTime)
			}
		}
		rule.Kind = RuleTypeScheduled
		rule.BlockedWindows = spec.Windows

	case spec.For > 0:
		rule.Kind = RuleTypeTimer
		rule.BlockedUntil = time.Now().Add(spec.For)

	case spec.Until != "":
		until, err := parseTimeOfDay(spec.Until)
		if err != nil {
			return DomainRule{}, fmt.Errorf("invalid --until %q: expected HH:MM", spec.Until)
		}
		// A time-of-day already past today means the user means the next
		// occurrence, so roll forward a day.
		if !until.After(time.Now()) {
			until = until.Add(24 * time.Hour)
		}
		rule.Kind = RuleTypeTimer
		rule.BlockedUntil = until

	default:
		rule.Kind = RuleTypeBlock
	}

	return rule, nil
}

type DomainsConfig struct {
	Rules map[string]DomainRule `toml:"rules"`
}

func LoadConfig(path string) ([]DomainRule, error) {
	var domainConfig DomainsConfig
	if _, err := toml.DecodeFile(path, &domainConfig); err != nil {
		return nil, err
	}

	rules := make([]DomainRule, 0, len(domainConfig.Rules))
	for domain, rule := range domainConfig.Rules {
		rule.Domain = domain
		rules = append(rules, rule)
	}
	return rules, nil
}

type AppendOutcome int

const (
	// OutcomeAdded: domain was new, rule created.
	OutcomeAdded AppendOutcome = iota
	// OutcomeWindowsMerged: scheduled rule existed, new window(s) appended.
	OutcomeWindowsMerged
	// OutcomeReplaced: an existing rule was overwritten via replace.
	OutcomeReplaced
	// OutcomeConflict: an incompatible rule already exists; nothing written.
	OutcomeConflict
)

type AppendResult struct {
	Outcome AppendOutcome
	// Rule is the resulting rule, or the attempted rule on a conflict.
	Rule DomainRule
	// Existing is the prior rule, set for merge/replace/conflict outcomes.
	Existing DomainRule
	// AddedWindows is the count of newly appended windows for a merge.
	AddedWindows int
}

// AppendRule adds newRule to the config. If the domain already exists, the
// outcome depends on the rule kinds and replace flag:
//   - replace=true: the existing rule is overwritten wholesale.
//   - scheduled existing + scheduled new: new windows are merged in (deduped).
//   - any other mismatch: OutcomeConflict, nothing is written.
func AppendRule(newRule DomainRule, replace bool) (AppendResult, error) {
	currentRules, err := LoadConfig(DomainsConfigPath)
	if err != nil {
		return AppendResult{}, err
	}

	idx := -1
	for i, r := range currentRules {
		if r.Domain == newRule.Domain {
			idx = i
			break
		}
	}

	if newRule.AddedAt.IsZero() {
		newRule.AddedAt = time.Now()
	}

	if idx == -1 {
		currentRules = append(currentRules, newRule)
		if err := writeConfig(DomainsConfigPath, currentRules); err != nil {
			return AppendResult{}, err
		}
		return AppendResult{Outcome: OutcomeAdded, Rule: newRule}, nil
	}

	existing := currentRules[idx]

	if replace {
		// Preserve when the domain was first restricted.
		newRule.AddedAt = existing.AddedAt
		currentRules[idx] = newRule
		if err := writeConfig(DomainsConfigPath, currentRules); err != nil {
			return AppendResult{}, err
		}
		return AppendResult{Outcome: OutcomeReplaced, Rule: newRule, Existing: existing}, nil
	}

	if existing.Kind == RuleTypeScheduled && newRule.Kind == RuleTypeScheduled {
		merged := existing
		added := 0
		for _, w := range newRule.BlockedWindows {
			if !slices.ContainsFunc(merged.BlockedWindows, w.equals) {
				merged.BlockedWindows = append(merged.BlockedWindows, w)
				added++
			}
		}
		if added == 0 {
			return AppendResult{Outcome: OutcomeWindowsMerged, Rule: existing, Existing: existing}, nil
		}
		currentRules[idx] = merged
		if err := writeConfig(DomainsConfigPath, currentRules); err != nil {
			return AppendResult{}, err
		}
		return AppendResult{Outcome: OutcomeWindowsMerged, Rule: merged, Existing: existing, AddedWindows: added}, nil
	}

	return AppendResult{Outcome: OutcomeConflict, Rule: newRule, Existing: existing}, nil
}

func RemoveRule(ruleDomain string) (string, error) {
	currentRules, err := LoadConfig(DomainsConfigPath)
	if err != nil {
		return "", err
	}

	ruleIndex := -1
	for index, currentRule := range currentRules {
		if ruleDomain == currentRule.Domain {
			ruleIndex = index
		}
	}
	if ruleIndex == -1 {
		return "", fmt.Errorf("Rule `%s` is not listed or already removed", ruleDomain)
	}
	updatedRules := slices.Delete(currentRules, ruleIndex, ruleIndex+1)
	if err := writeConfig(DomainsConfigPath, updatedRules); err != nil {
		return "", err
	}
	return ruleDomain, nil
}

func writeConfig(path string, rules []DomainRule) error {
	sorted := make([]DomainRule, len(rules))
	copy(sorted, rules)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Domain < sorted[j].Domain
	})

	var sb strings.Builder
	sb.WriteString("[rules]\n")
	for _, rule := range sorted {
		sb.WriteString(rule.ConfigStr() + "\n")
	}
	return os.WriteFile(path, []byte(sb.String()), 0644)
}
