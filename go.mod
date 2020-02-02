module github.com/terraform-providers/terraform-provider-pagerduty

require (
	github.com/PagerDuty/go-pagerduty v0.0.0-00010101000000-000000000000
	github.com/hashicorp/terraform-plugin-sdk v1.0.0
	github.com/heimweh/go-pagerduty v0.0.0-20190807171021-2a6540956dc5 // indirect
)

replace github.com/PagerDuty/go-pagerduty => ../../PagerDuty/go-pagerduty

go 1.13
