package config

type Config struct {
	ProfilePath  string
	OutputFormat string // human, json, sarif
	FailOn       string // warn, block
	Paths        []string
	Stdin        bool
}
