module github.com/gklsndr/kolavatar-playground/go

go 1.25.0

require github.com/gklsndr/kolavatar v0.0.0

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/klauspost/cpuid/v2 v2.0.9 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/prometheus/client_golang v1.23.2 // indirect
	github.com/prometheus/client_model v0.6.2 // indirect
	github.com/prometheus/common v0.66.1 // indirect
	github.com/prometheus/procfs v0.16.1 // indirect
	go.opentelemetry.io/otel v1.43.0 // indirect
	go.opentelemetry.io/otel/trace v1.43.0 // indirect
	go.yaml.in/yaml/v2 v2.4.2 // indirect
	golang.org/x/sys v0.35.0 // indirect
	google.golang.org/protobuf v1.36.8 // indirect
	lukechampine.com/blake3 v1.4.1 // indirect
)

// Local sibling checkout. Once kolavatar is tagged on GitHub, drop the replace
// and bump the require line to a real version (e.g. v1.0.0).
replace github.com/gklsndr/kolavatar => ../../kolavatar-go
