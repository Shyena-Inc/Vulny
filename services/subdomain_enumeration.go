package services

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

type SubdomainResult struct {
	Subdomain string   `json:"subdomain" bson:"subdomain"`
	IPs       []string `json:"ips" bson:"ips"`
}
 
func EnumerateSubdomainsFromFileConcurrent(domain string, filepath string) ([]SubdomainResult, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to open wordlist file: %w", err)
	}
	defer file.Close()

	var wg sync.WaitGroup
	resultsCh := make(chan SubdomainResult)
	errCh := make(chan error, 1)

	concurrencyLimit := 20
	sem := make(chan struct{}, concurrencyLimit)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Read all subdomains into a slice first
	var subs []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		sub := strings.TrimSpace(scanner.Text())
		if sub != "" {
			subs = append(subs, sub)
		}
	}
	if scannerErr := scanner.Err(); scannerErr != nil {
		return nil, fmt.Errorf("error reading wordlist file: %w", scannerErr)
	}

	results := []SubdomainResult{}
 
	for _, sub := range subs {
		wg.Add(1)
		go func(sub string) {
			defer wg.Done()
			sem <- struct{}{} // acquire semaphore
			defer func() { <-sem }() // release semaphore

			fullDomain := fmt.Sprintf("%s.%s", sub, domain)
			ips, err := net.DefaultResolver.LookupHost(ctx, fullDomain)
			if err == nil && len(ips) > 0 {
				resultsCh <- SubdomainResult{
					Subdomain: fullDomain,
					IPs:       ips,
				}
			}
		}(sub)
	}
 
	go func() {
		wg.Wait()
		close(resultsCh)
	}()
 
	for r := range resultsCh {
		results = append(results, r)
	}

	select {
	case e := <-errCh:
		return nil, e
	default:
	}

	return results, nil
}
