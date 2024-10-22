package apis

// APIResponse represents the structure of the response to extract API IDs.
type APIResponse struct {
	ID string `json:"id"`
}

// RevisionResponse represents the structure of the revision response.
type RevisionResponse struct {
	ID      string `json:"id"`
	ApiInfo struct {
		ID string `json:"id"`
	} `json:"apiInfo"`
}
