package apis

import (
	"crypto/tls"
	"net/http"
)

const (
	apisBasePath       = "https://localhost:9444/api/am/publisher/v2/apis"
	envsBasePath       = "https://localhost:9444/api/am/admin/v2/environments"
	dataplanesBasePath = "https://localhost:9444/api/choreo/internal/v1/dataplanes"
	revisionPath       = "%s/%s/revisions"
	openAPIVersion     = "v3"
)

var insecureClient = &http.Client{
	Transport: &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	},
}
