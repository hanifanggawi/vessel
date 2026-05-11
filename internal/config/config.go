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

func AppendRule(rules []DomainRule) ([]DomainRule, error) {
	currentRules, err := LoadConfig(DomainsConfigPath)
	if err != nil {
		return nil, err
	}

	ruleMap := make(map[string]DomainRule, len(currentRules)+len(rules))
	for _, rule := range currentRules {
		ruleMap[rule.Domain] = rule
	}

	now := time.Now()
	for _, newRule := range rules {
		if _, exists := ruleMap[newRule.Domain]; exists {
			continue
		}
		if newRule.AddedAt.IsZero() {
			newRule.AddedAt = now
		}
		ruleMap[newRule.Domain] = newRule
		currentRules = append(currentRules, newRule)
	}

	if err := writeConfig(DomainsConfigPath, currentRules); err != nil {
		return nil, err
	}
	return currentRules, nil
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
