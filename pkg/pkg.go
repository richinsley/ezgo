package pkg

import (
	"log"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

func HandlePkgCommand(args []string, isQuiet bool) {
	if len(args) == 0 {
		log.Fatalln("Usage: ezgo pkg <add|tidy> [arguments]")
	}
	subcommand, subcommandArgs := args[0], args[1:]

	cfg, err := LoadProjectConfig()
	if err != nil {
		log.Fatalf("!!! ezgo: Error reading .ezgo.yml: %v", err)
	}
	if cfg == nil {
		log.Fatalln("!!! ezgo: .ezgo.yml not found. Run 'ezgo mod init' first.")
	}

	switch subcommand {
	case "add":
		if len(subcommandArgs) == 0 {
			log.Fatalln("Usage: ezgo pkg add <package1> [package2]...")
		}
		existing := make(map[string]bool)
		for _, pkg := range cfg.Packages {
			existing[pkg] = true
		}
		var packagesAdded []string
		for _, newPkg := range subcommandArgs {
			if !existing[newPkg] {
				cfg.Packages = append(cfg.Packages, newPkg)
				existing[newPkg] = true
				packagesAdded = append(packagesAdded, newPkg)
			}
		}

		if len(packagesAdded) == 0 {
			log.Println("ezgo: all specified packages already exist in .ezgo.yml.")
			return
		}
		log.Printf("ezgo: added %s to .ezgo.yml", strings.Join(packagesAdded, ", "))

		data, err := yaml.Marshal(cfg)
		if err != nil {
			log.Fatalf("!!! ezgo: Failed to serialize config: %v", err)
		}
		if err := os.WriteFile(".ezgo.yml", data, 0644); err != nil {
			log.Fatalf("!!! ezgo: Failed to write .ezgo.yml: %v", err)
		}

		// After adding, ensure the new packages are installed
		_, err = SetupCGOEnvironment(isQuiet, packagesAdded)
		if err != nil {
			log.Fatalf("!!! ezgo: Failed to install new packages: %v", err)
		}
		if !isQuiet {
			log.Println("--- ezgo: Environment updated successfully.")
		}

	case "tidy":
		if !isQuiet {
			log.Println("--- ezgo: Tidying environment...")
		}
		_, err = SetupCGOEnvironment(isQuiet, cfg.Packages)
		if err != nil {
			log.Fatalf("!!! ezgo: Failed to sync environment with .ezgo.yml: %v", err)
		}
		if !isQuiet {
			log.Println("--- ezgo: Environment is up to date.")
		}

	default:
		log.Fatalf("Unknown command: 'ezgo pkg %s'. Use 'add' or 'tidy'.", subcommand)
	}
}
