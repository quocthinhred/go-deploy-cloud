package model

const (
	STG = "stg"
	DEV = "dev"
)

type Account struct {
	Username   string `json:"username"`
	Password   string `json:"password"`
	Type       string `json:"type"`
	DomainType string `json:"domainType"`
}
