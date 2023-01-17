// Once In-Use Image data has been gathered, this package reports the data to xeol.io
package reporter

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/noqcks/xeol-agent/agent/inventory"
	"github.com/noqcks/xeol-agent/internal/config"
	"github.com/noqcks/xeol-agent/internal/log"
)

const ReportAPIPath = "v1/inventories"
const ReportAPIURL = "https://engine.xeol.io/" + ReportAPIPath

// This method does the actual Reporting (via HTTP) to xeol.io
//
//nolint:gosec
func Put(report inventory.Report, xeolDetails config.XeolInfo, appConfig *config.Application) error {
	log.Debug("Reporting results to engine.xeol.io")
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: xeolDetails.HTTP.Insecure},
	}
	client := &http.Client{
		Transport: tr,
		Timeout:   time.Duration(xeolDetails.HTTP.TimeoutSeconds) * time.Second,
	}

	reqBody, err := json.Marshal(report)
	if err != nil {
		return fmt.Errorf("failed to serialize results as JSON: %w", err)
	}

	req, err := http.NewRequest("PUT", ReportAPIURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return fmt.Errorf("failed to build request to report data to xeol.io: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("ApiKey %s", xeolDetails.APIKey))
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to report data to xeol.io: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("failed to report data to xeol.io: %+v", resp)
	}
	log.Debug("Successfully reported results to xeol.io")
	return nil
}
