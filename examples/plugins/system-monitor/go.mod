module nixai-plugin-system-monitor

go 1.23.0

toolchain go1.24.4

require nix-ai-help v0.0.0

require (
	github.com/kr/text v0.2.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace nix-ai-help => ../../../
