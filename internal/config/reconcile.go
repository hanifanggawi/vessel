package config

import (
	"fmt"
	"os"
	"time"

	"github.com/hanifanggawi/vessel/internal/hosts"
)

const (
	redirectIP = "0.0.0.0"
)

var defaultSubdomains = []string{
	"",
	"www.",
}

func RunReconcile() error {
	// A missing config means nothing should be blocked. The boot-time daemon
	// reconciles every 60s and would otherwise log a decode error each tick
	// before `vessel init` has run; treat absence as an empty ruleset so the
	// managed block is simply cleared.
	if _, err := os.Stat(DomainsConfigPath); os.IsNotExist(err) {
		return reconcileHostsFile(nil, EtcHostsPath)
	}

	domainRules, err := LoadConfig(DomainsConfigPath)
	if err != nil {
		return err
	}
	err = reconcileHostsFile(domainRules, EtcHostsPath)
	if err != nil {
		return err
	}
	return nil
}

func reconcileHostsFile(rules []DomainRule, hostsPath string) error {
	// check hostsPaths exists
	_, err := os.Stat(hostsPath)
	if err != nil {
		return err
	}
	// for every rule generate the hosts entry, adjust for each rule's kind
	hostsEntries := make([]string, 0)
	for _, rule := range rules {
		switch rule.Kind {
		case RuleTypeBlock:
			hostsEntries = append(hostsEntries, generateDomainEntries(rule.Domain)...)
		case RuleTypeScheduled:
			// validate schedule, append entry only if time now fits within schedule
			addEntry := false
			for _, window := range rule.BlockedWindows {
				start, err := window.ParseStart()
				end, err := window.ParseEnd()
				now := time.Now()
				if err != nil {
					return err
				}
				if now.After(start) && now.Before(end) {
					addEntry = true
					break
				}
			}
			if addEntry {
				hostsEntries = append(hostsEntries, generateDomainEntries(rule.Domain)...)
			}
		case RuleTypeTimer:
			// validate timer, append entry only if time now is <= block_until
			if time.Now().Before(rule.BlockedUntil) {
				fmt.Println("")
				hostsEntries = append(hostsEntries, generateDomainEntries(rule.Domain)...)
			}
		}
	}

	err = hosts.UpdateHostsEntries(hostsEntries)
	if err != nil {
		return err
	}

	return nil
}

func generateDomainEntries(domain string) []string {
	entries := make([]string, 0, len(defaultSubdomains))
	for _, subdomain := range defaultSubdomains {
		entries = append(entries, fmt.Sprintf("%s %s%s", redirectIP, subdomain, domain))
	}
	return entries
}
