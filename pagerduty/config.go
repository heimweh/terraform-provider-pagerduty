package pagerduty

import (
	"fmt"
	"log"
	"net/http"

	"github.com/PagerDuty/go-pagerduty"
	"github.com/hashicorp/terraform-plugin-sdk/helper/logging"
)

// Config defines the configuration options for the PagerDuty client
type Config struct {
	// The PagerDuty API V2 token
	Token string

	// Skip validation of the token against the PagerDuty API
	SkipCredsValidation bool

	// UserAgent for API Client
	UserAgent string
}

const invalidCreds = `

No valid credentials found for PagerDuty provider.
Please see https://www.terraform.io/docs/providers/pagerduty/index.html
for more information on providing credentials for this provider.
`

// Client returns a new PagerDuty client
func (c *Config) Client() (*pagerduty.Client, error) {
	// Validate that the PagerDuty token is set
	if c.Token == "" {
		return nil, fmt.Errorf(invalidCreds)
	}

	var httpClient *http.Client
	httpClient = http.DefaultClient
	httpClient.Transport = logging.NewTransport("PagerDuty", http.DefaultTransport)

	client := pagerduty.NewClient(c.Token)

	if !c.SkipCredsValidation {
		// Validate the credentials by calling the abilities endpoint,
		// if we get a 401 response back we return an error to the user
		if _, err := client.ListAbilities(); err != nil {
			return nil, fmt.Errorf(fmt.Sprintf("%s\n%s", err, invalidCreds))
		}
	}

	log.Printf("[INFO] PagerDuty client configured")

	return client, nil
}
