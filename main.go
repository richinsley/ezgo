package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"

	ezgo "github.com/richinsley/ezgo/pkg"
)

func main() {
	log.SetFlags(0)

	if len(os.Args) < 2 {
		fmt.Println("ezgo: A CGO-aware wrapper for the Go compiler on Windows.")
		fmt.Println("\nUsage:")
		fmt.Println("  ezgo [-q] <go_command> [arguments]")
		fmt.Println("  ezgo build [-no-copy] [arguments]")
		fmt.Println("  ezgo env <clean|path|vars>")
		fmt.Println("  ezgo mod init")
		fmt.Println("  ezgo pkg <add|tidy> [packages...]")
		fmt.Println("  ezgo shell [powershell|cmd]")
		fmt.Println("\nExamples:")
		fmt.Println("  ezgo build -o myapp.exe .")
		fmt.Println("  ezgo pkg add glfw")
		os.Exit(1)
	}

	args := os.Args[1:]
	isQuiet := false
	if len(args) > 0 && args[0] == "-q" {
		isQuiet = true
		args = args[1:]
	}

	if len(args) == 0 {
		main() // Show help
		return
	}

	switch args[0] {
	case "env":
		ezgo.HandleEnvCommand(args[1:], isQuiet)
		return
	case "mod":
		ezgo.HandleModCommand(args[1:])
		return
	case "pkg":
		ezgo.HandlePkgCommand(args[1:], isQuiet)
		return
	case "shell":
		ezgo.HandleShellCommand(args[1:], isQuiet)
		return
	}

	projectCfg, err := ezgo.LoadProjectConfig()
	if err != nil {
		log.Fatalf("!!! ezgo: Error reading .ezgo.yml: %v", err)
	}

	// For build commands, we only ensure the base environment exists.
	// We DO NOT install packages. 'ezgo pkg tidy' is for that.
	cgoEnv, err := ezgo.SetupCGOEnvironment(isQuiet, nil)
	if err != nil {
		log.Fatalf("!!! ezgo: Failed to configure CGO environment: %v", err)
	}

	goExecutable, err := exec.LookPath("go")
	if err != nil {
		log.Fatalf("!!! ezgo: Could not find 'go' executable in your system's PATH: %v", err)
	}

	// --- Handle build-specific flags ---
	finalArgs := args
	isBuildCommand := args[0] == "build"
	skipCopy := false
	if isBuildCommand {
		var filteredArgs []string
		for _, arg := range args {
			if arg == "-no-copy" {
				skipCopy = true
			} else {
				filteredArgs = append(filteredArgs, arg)
			}
		}
		finalArgs = filteredArgs
	}

	cmd := exec.Command(goExecutable, finalArgs...)
	cmd.Env = cgoEnv.GetGoEnv(projectCfg)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	// Explicitly set the working directory for the 'go' command.
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("!!! ezgo: could not get current working directory: %v", err)
	}
	cmd.Dir = cwd

	err = cmd.Run()

	// After a successful build, copy the necessary runtime DLLs.
	if err == nil && isBuildCommand {
		if err := ezgo.HandlePostBuild(finalArgs, cgoEnv, isQuiet, skipCopy); err != nil {
			log.Fatalf("!!! ezgo: Post-build step failed: %v", err)
		}
	}

	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			os.Exit(exitError.ExitCode())
		} else {
			log.Fatalf("!!! ezgo: Command failed with an unknown error: %v", err)
		}
	}
}
