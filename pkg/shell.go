package pkg

import (
	"log"
	"os"
	"os/exec"
	"strings"
)

func HandleShellCommand(args []string, isQuiet bool) {
	// Get the configured CGO environment object. This ensures micromamba exists.
	cgoEnv, err := SetupCGOEnvironment(isQuiet, nil)
	if err != nil {
		log.Fatalf("!!! ezgo: could not setup environment for shell command: %v", err)
	}

	// Determine which shell executable to use.
	shellExe := "cmd.exe" // Default to cmd
	if len(args) > 0 && strings.ToLower(args[0]) == "powershell" {
		shellExe = "powershell.exe"
	}

	// Find the full path to the shell executable.
	shellPath, err := exec.LookPath(shellExe)
	if err != nil {
		log.Fatalf("!!! ezgo: could not find '%s' in your system's PATH: %v", shellExe, err)
	}

	if !isQuiet {
		log.Printf("--- ezgo: Starting interactive %s shell with CGO environment...", shellExe)
	}

	// Prepare the command.
	cmd := exec.Command(shellPath)

	// Get the full set of environment variables.
	// We pass nil for the project config to get a general-purpose CGO shell.
	cmd.Env = cgoEnv.GetGoEnv(nil)

	// Connect Stdin, Stdout, and Stderr to make the shell interactive in the current window.
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Run the shell. This will block until the user types 'exit'.
	if err := cmd.Run(); err != nil {
		log.Fatalf("!!! ezgo: failed to start interactive shell: %v", err)
	}

	if !isQuiet {
		log.Println("--- ezgo: Shell session ended.")
	}
}
