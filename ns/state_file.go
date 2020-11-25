package ns

type StateFile struct {
	Version          int     `json:"version"`
	TerraformVersion string  `json:"terraform_version"`
	Serial           int64   `json:"serial"`
	Lineage          string  `json:"lineage"`
	Outputs          Outputs `json:"outputs"`
}
