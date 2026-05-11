package daemon

import (
	"fmt"
	"os"

	"github.com/hanifanggawi/vessel/internal/config"
	. "github.com/hanifanggawi/vessel/internal/config"
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
	domainRules, err := config.LoadConfig(DomainsConfigPath)
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
			// TODO: validate schedule, append entry only if time now fits within schedule
			hostsEntries = append(hostsEntries, generateDomainEntries(rule.Domain)...)
		case RuleTypeTimer:
			// TODO: validate timer, append entry only if time now is <= block_until
			hostsEntries = append(hostsEntries, generateDomainEntries(rule.Domain)...)
		}
	}

	for _, entry := range hostsEntries {
		fmt.Printf("DISINI entry: %+v\n", entry)
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
		// if subdomain == "" {
		// 	entries = append(entries, "%s %s", redirectIP, domain)
		// }
		entries = append(entries, fmt.Sprintf("%s %s%s", redirectIP, subdomain, domain))
	}
	return entries
}
