// Once In-Use Image data has been gathered, this package reports the data to Anchore
package reporter

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/anchore/kai/internal/config"
	"github.com/anchore/kai/internal/log"
	"github.com/anchore/kai/kai/inventory"
)

const ReportAPIPath = "v1/enterprise/inventories"

// This method does the actual Reporting (via HTTP) to Anchore
//
//nolint:gosec
func Post(report inventory.Report, XeolDetails config.XeolInfo, appConfig *config.Application) error {
	log.Debug("Reporting results to Anchore")
	// tr := &http.Transport{
	// 	TLSClientConfig: &tls.Config{InsecureSkipVerify: XeolDetails.HTTP.Insecure},
	// }
	// client := &http.Client{
	// 	Transport: tr,
	// 	Timeout:   time.Duration(XeolDetails.HTTP.TimeoutSeconds) * time.Second,
	// }

	// anchoreURL, err := buildURL(XeolDetails)
	// if err != nil {
	// return fmt.Errorf("failed to build url: %w", err)
	// }

	// anchoreURL := "https://xeol.io/api/v1/enterprise/inventories"

	reqBody, err := json.Marshal(report)
	if err != nil {
		return fmt.Errorf("failed to serialize results as JSON: %w", err)
	}

	b, _ := prettyprint(reqBody)
	fmt.Printf("%s", b)

	// req, err := http.NewRequest("POST", anchoreURL, bytes.NewBuffer(reqBody))
	// if err != nil {
	// 	return fmt.Errorf("failed to build request to report data to Anchore: %w", err)
	// }
	// // TODO(benji): update authentication to xeol.io backend
	// // req.SetBasicAuth(XeolDetails.User, XeolDetails.Password)
	// req.Header.Set("Content-Type", "application/json")
	// // req.Header.Set("x-anchore-account", XeolDetails.Account)
	// resp, err := client.Do(req)
	// if err != nil {
	// 	return fmt.Errorf("failed to report data to Anchore: %w", err)
	// }
	// defer resp.Body.Close()
	// if resp.StatusCode != 200 {
	// 	return fmt.Errorf("failed to report data to Anchore: %+v", resp)
	// }
	log.Debug("Successfully reported results to xeol.io")
	return nil
}

// func buildURL(XeolDetails config.XeolInfo) (string, error) {
// 	anchoreURL, err := url.Parse(XeolDetails.URL)
// 	if err != nil {
// 		return "", err
// 	}

// 	anchoreURL.Path += ReportAPIPath

// 	return anchoreURL.String(), nil
// }

func prettyprint(b []byte) ([]byte, error) {
	var out bytes.Buffer
	err := json.Indent(&out, b, "", "  ")
	return out.Bytes(), err
}
