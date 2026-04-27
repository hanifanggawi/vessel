package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type RuleType string

const (
	RuleTypeBlock         RuleType = "block"
	RuleTypeTimer         RuleType = "timer"
	RuleTypeDurationLimit RuleType = "duration_limit"
)

type DomainRule struct {
	domain       string
	kind         RuleType
	addedAt      time.Time
	blockedUntil time.Time
}

func getExamleConfigPath() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return filepath.Join(cwd, "bin", "hosts_test"), nil
}

func LoadConfig(path string) ([]DomainRule, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	if len(data) <= 0 {
		fmt.Println("Filler aja")
	}

	// parse data
	var rules []DomainRule

	return rules, nil
}
