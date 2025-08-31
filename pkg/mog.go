package pkg

import (
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

func HandleModCommand(args []string) {
	if len(args) == 0 || args[0] != "init" {
		log.Fatalln("Usage: ezgo mod init")
	}
	configFile := ".ezgo.yml"
	if _, err := os.Stat(configFile); err == nil {
		log.Fatalf("!!! ezgo: %s already exists.", configFile)
	}

	defaultConfig := ProjectConfig{
		Packages:    []string{},
		Environment: map[string]string{"YOUR_VAR": "your_value"},
	}
	data, err := yaml.Marshal(&defaultConfig)
	if err != nil {
		log.Fatalf("!!! ezgo: Failed to create default config: %v", err)
	}

	comment := []byte(`# Add conda-forge package names for your CGO project
#
# packages:
#   - glfw
#
# Add custom environment variables to be passed to the go compiler
#
# environment:
#   SOME_FLAG: "true"
`)
	data = append(comment, data...)

	if err := os.WriteFile(configFile, data, 0644); err != nil {
		log.Fatalf("!!! ezgo: Failed to write %s: %v", configFile, err)
	}
	log.Printf("ezgo: created %s", configFile)
}
