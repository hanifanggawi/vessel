package hosts

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"

	mapset "github.com/deckarep/golang-set/v2"
)

const (
	beginMarker = "#---BEGIN VESSEL MANAGED DOMAINS---#"
	endMarker   = "#---END VESSEL MANAGED DOMAINS-----#"
	redirectIP  = "0.0.0.0"
)

var defaultSubdomains = []string{
	"",
	// "www.",
}

func GetHostsPath() (string, error) {
	// cwd, err := os.Getwd()
	// if err != nil {
	// 	return "", err
	// }
	// return filepath.Join(cwd, "bin", "hosts_test"), nil
	switch runtime.GOOS {
	case "windows":
		return filepath.Join(os.Getenv("SystemRoot"), "System32", "drivers", "etc", "hosts"), nil
	case "linux":
		return "/etc/hosts", nil
	default:
		return "", fmt.Errorf("Unsupported OS %s", runtime.GOOS)
	}
}

func getDomains() []string {
	hostsPath, err := GetHostsPath()
	if err != nil {
		fmt.Println(err.Error())
	}
	data, err := os.ReadFile(hostsPath)
	if err != nil {
		fmt.Println(err.Error())
	}

	lines := strings.Split(string(data), "\n")
	read := false
	var domains []string
	for _, line := range lines {
		// find marker
		if line == beginMarker {
			read = true
			continue
		}
		if !read {
			continue
		}
		if line == endMarker {
			break
		}
		domain := strings.Fields(line)[1]
		if !slices.Contains(domains, domain) {
			domains = append(domains, domain)
		}
	}
	return domains
}

func getDomainsSet() mapset.Set[string] {
	hostsPath, err := GetHostsPath()
	if err != nil {
		fmt.Println(err.Error())
	}
	data, err := os.ReadFile(hostsPath)
	if err != nil {
		fmt.Println(err.Error())
	}

	lines := strings.Split(string(data), "\n")
	read := false
	domains := mapset.NewSet[string]()
	for _, line := range lines {
		// find marker
		if line == beginMarker {
			read = true
			continue
		}
		if !read {
			continue
		}
		if line == endMarker {
			break
		}
		domain := strings.Fields(line)[1]
		domains.Add(domain)
	}
	return domains
}

func UpdateHostsEntries(entries []string) error {
	hostsPath, err := GetHostsPath()
	if err != nil {
		fmt.Println(err.Error())
	}
	data, err := os.ReadFile(hostsPath)
	if err != nil {
		return err
	}

	lines := strings.Split(string(data), "\n")
	var beginIndex int
	var endIndex int
	for index, line := range lines {
		if line == beginMarker {
			beginIndex = index
		}
		if line == endMarker {
			endIndex = index
		}
	}
	updatedLines := make([]string, 0, len(lines))
	for i := 0; i <= beginIndex; i++ {
		updatedLines = append(updatedLines, lines[i])
	}
	for _, ruleEntry := range entries {
		updatedLines = append(updatedLines, ruleEntry)
	}
	// for domain := range mapset.Elements(domains) {
	// 	for _, subdomain := range defaultSubdomains {
	// 		if subdomain == "" {
	// 			updatedLines = append(updatedLines, fmt.Sprintf("%s %s", redirectIP, domain))
	// 		} else {
	// 			updatedLines = append(updatedLines, fmt.Sprintf("%s %s.%s", redirectIP, subdomain, domain))
	// 		}
	// 	}
	// }
	for j := endIndex; j < len(lines); j++ {
		updatedLines = append(updatedLines, lines[j])
	}

	newContent := strings.Join(updatedLines, "\n")

	outputPath, err := GetHostsPath()
	if err != nil {
		return err
	}

	// fmt.Printf("\nExpected new content %s\n", newContent)

	err = os.WriteFile(outputPath, []byte(newContent), 0644)
	if err != nil {
		if os.IsPermission(err) {
			switch runtime.GOOS {
			case "linux":
				return fmt.Errorf("permission denied writing to %s: try running with sudo", outputPath)
			case "windows":
				return fmt.Errorf("permission denied writing to %s: try running as Administrator", outputPath)
			}
		}
		return err
	}

	return nil
}

func updateDomains(domains mapset.Set[string]) error {
	hostsPath, err := GetHostsPath()
	if err != nil {
		fmt.Println(err.Error())
	}
	data, err := os.ReadFile(hostsPath)
	if err != nil {
		return err
	}

	lines := strings.Split(string(data), "\n")
	var beginIndex int
	var endIndex int
	for index, line := range lines {
		if line == beginMarker {
			beginIndex = index
		}
		if line == endMarker {
			endIndex = index
		}
	}
	updatedLines := make([]string, 0, len(lines))
	for i := 0; i <= beginIndex; i++ {
		updatedLines = append(updatedLines, lines[i])
	}
	for domain := range mapset.Elements(domains) {
		for _, subdomain := range defaultSubdomains {
			if subdomain == "" {
				updatedLines = append(updatedLines, fmt.Sprintf("%s %s", redirectIP, domain))
			} else {
				updatedLines = append(updatedLines, fmt.Sprintf("%s %s.%s", redirectIP, subdomain, domain))
			}
		}
	}
	for j := endIndex; j < len(lines); j++ {
		updatedLines = append(updatedLines, lines[j])
	}

	newContent := strings.Join(updatedLines, "\n")

	outputPath, err := GetHostsPath()
	if err != nil {
		return err
	}

	// fmt.Printf("\nExpected new content %s\n", newContent)

	err = os.WriteFile(outputPath, []byte(newContent), 0644)
	if err != nil {
		if os.IsPermission(err) {
			switch runtime.GOOS {
			case "linux":
				return fmt.Errorf("permission denied writing to %s: try running with sudo", outputPath)
			case "windows":
				return fmt.Errorf("permission denied writing to %s: try running as Administrator", outputPath)
			}
		}
		return err
	}

	return nil
}

func checkBoundaryExists() bool {
	hostsPath, err := GetHostsPath()
	if err != nil {
		fmt.Println(err.Error())
	}
	data, err := os.ReadFile(hostsPath)
	if err != nil {
		fmt.Println(err.Error())
	}

	lines := strings.Split(string(data), "\n")
	beginMarkerExists, endMarkerExists := false, false
	for _, line := range lines {
		// find marker
		switch line {
		case beginMarker:
			beginMarkerExists = true
		case endMarker:
			endMarkerExists = true
		}
	}
	return beginMarkerExists && endMarkerExists
}

func insertBoundaryMarkers() error {
	hostsPath, err := GetHostsPath()
	if err != nil {
		return err
	}

	outputPath := hostsPath
	data, err := os.ReadFile(hostsPath)
	if err != nil {
		return err
	}
	lines := strings.Split(string(data), "\n")
	lines = append(lines, "\n")
	lines = append(lines, beginMarker)
	lines = append(lines, endMarker)
	newContent := strings.Join(lines, "\n")

	err = os.WriteFile(outputPath, []byte(newContent), 0644)
	if err != nil {
		if os.IsPermission(err) {
			return fmt.Errorf("permission denied writing to %s: try running with sudo", outputPath)
		}
		return err
	}

	return nil
}

func Add(domain string) error {
	if !checkBoundaryExists() {
		err := insertBoundaryMarkers()
		if err != nil {
			return err
		}
	}
	domains := getDomainsSet()
	if domains.Contains(domain) {
		fmt.Printf("%s is already listed", domain)
		return nil
	}
	domains.Add(domain)
	err := updateDomains(domains)
	if err != nil {
		return err
	}
	return nil
}

func List() {
	domains := getDomains()
	for _, domain := range domains {
		fmt.Println(domain)
	}
	fmt.Println()
}

func Remove(domain string) error {
	domains := getDomainsSet()

	for domain := range domains.Iter() {
		fmt.Printf("DISINI domain: %+v\n", domain)
	}
	var domainsToRemove []string
	for _, subdomain := range defaultSubdomains {
		ok := domains.ContainsOne(domain)
		if ok {
			domains.Remove(subdomain + domain)
			domainsToRemove = append(domainsToRemove, subdomain+domain)
		}
	}
	if len(domainsToRemove) == 0 {
		fmt.Printf("'%s' domain not listed or already removed", domain)
		return nil
	}

	err := updateDomains(domains)

	if err != nil {
		return err
	}

	fmt.Println("Removed entries")
	for _, entry := range domainsToRemove {
		fmt.Println(entry)
	}

	return nil
}
