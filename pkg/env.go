package pkg

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	jumpboot "github.com/richinsley/jumpboot"
)

// SetupCGOEnvironment ensures the CGO toolchain and any specified packages exist.
func SetupCGOEnvironment(isQuiet bool, packagesToInstall []string) (*CGOEnvironment, error) {
	cacheRoot, err := GetCacheRoot()
	if err != nil {
		return nil, err
	}

	// Use the main cache root for jumpboot. This ensures micromamba.exe is in a predictable location.
	rootDirectory := cacheRoot

	// Define the environment details.
	envName := "cgo_win_env_py312"
	pythonVersion := "3.12"
	channel := "conda-forge"
	envDirectory := filepath.Join(rootDirectory, "envs", envName)

	if _, err := os.Stat(envDirectory); err != nil && !isQuiet {
		log.Println("--- ezgo: First run detected. Setting up CGO toolchain (this may take a few minutes)...")
	}

	env, err := jumpboot.CreateEnvironmentMamba(envName, rootDirectory, pythonVersion, channel, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating environment: %w", err)
	}

	if env.IsNew {
		if !isQuiet {
			log.Printf("--- ezgo: Installing %s...", "m2w64-toolchain_win-64")
		}
		if err := env.MicromambaInstallPackage("m2w64-toolchain_win-64", channel); err != nil {
			return nil, fmt.Errorf("error installing package m2w64-toolchain_win-64: %w", err)
		}
		if !isQuiet {
			log.Println("--- ezgo: Toolchain installation complete.")
		}
	}

	if len(packagesToInstall) > 0 {
		if !isQuiet {
			log.Println("--- ezgo: Ensuring project-specific dependencies are installed...")
		}
		for _, pkg := range packagesToInstall {
			if err := env.MicromambaInstallPackage(pkg, channel); err != nil {
				return nil, fmt.Errorf("error installing project package %s: %w", pkg, err)
			}
		}
	}

	mingwRoot := filepath.Join(env.EnvPath, "Library", "mingw-w64")
	triplet := "x86_64-w64-mingw32"

	cgoenv := &CGOEnvironment{
		Environment:     env,
		CompilerBinPath: filepath.Join(mingwRoot, "bin"),
		ToolBinPath:     filepath.Join(mingwRoot, triplet, "bin"),
		IncludePath:     filepath.Join(mingwRoot, triplet, "include"),
		LibPath:         filepath.Join(mingwRoot, triplet, "lib"),
	}
	// FullBinPath will be fully constructed in GetGoEnv to include all necessary paths.
	cgoenv.FullBinPath = fmt.Sprintf("%s;%s", cgoenv.CompilerBinPath, cgoenv.ToolBinPath)

	return cgoenv, nil
}

// GetCacheRoot returns the root directory for all ezgo cache and environment data.
func GetCacheRoot() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not get user home directory: %w", err)
	}
	return filepath.Join(homeDir, ".cache", "ezgo"), nil
}

// GetGoEnv returns a slice of strings representing the environment variables for CGO.
func (c *CGOEnvironment) GetGoEnv(cfg *ProjectConfig) []string {
	// Start with the current system environment
	env := os.Environ()
	envMap := make(map[string]string)
	for _, e := range env {
		parts := strings.SplitN(e, "=", 2)
		if len(parts) == 2 {
			envMap[strings.ToUpper(parts[0])] = parts[1]
		}
	}

	// Apply custom environment variables from .ezgo.yml
	if cfg != nil && cfg.Environment != nil {
		for k, v := range cfg.Environment {
			envMap[strings.ToUpper(k)] = v
		}
	}

	// Define general environment paths for packages
	generalLibRoot := filepath.Join(c.EnvPath, "Library")
	generalIncludePath := filepath.Join(generalLibRoot, "include")
	generalLibPath := filepath.Join(generalLibRoot, "lib")
	generalBinPath := filepath.Join(generalLibRoot, "bin")

	// Construct the full, correct PATH, CFLAGS, and LDFLAGS
	// Order matters: prioritize the specific compiler, then general package bins, then toolchain DLLs, then the env root (for python, etc)
	fullPath := strings.Join([]string{c.CompilerBinPath, generalBinPath, c.ToolBinPath, c.EnvPath, os.Getenv("PATH")}, ";")
	// Search for headers in the general package dir first, then the core toolchain dir.
	cgoCFLAGS := fmt.Sprintf("-I%s -I%s", generalIncludePath, c.IncludePath)
	// Search for libs in the general package dir first, then the core toolchain dir.
	cgoLDFLAGS := fmt.Sprintf("-L%s -L%s", generalLibPath, c.LibPath)

	// Apply the core CGO variables, overriding any custom or system variables
	cgoVars := map[string]string{
		"CGO_ENABLED": "1",
		"CC":          filepath.Join(c.CompilerBinPath, "gcc.exe"),
		"CXX":         filepath.Join(c.CompilerBinPath, "g++.exe"),
		"PATH":        fullPath,
		"CGO_CFLAGS":  cgoCFLAGS,
		"CGO_LDFLAGS": cgoLDFLAGS,
	}

	for k, v := range cgoVars {
		envMap[strings.ToUpper(k)] = v
	}

	var newEnv []string
	for k, v := range envMap {
		newEnv = append(newEnv, fmt.Sprintf("%s=%s", k, v))
	}
	return newEnv
}

func HandleEnvCommand(args []string, isQuiet bool) {
	if len(args) == 0 {
		log.Fatalln("Usage: ezgo env <clean|path|vars>")
	}
	cacheRoot, err := GetCacheRoot()
	if err != nil {
		log.Fatalf("!!! ezgo: %v", err)
	}

	switch args[0] {
	case "clean":
		if _, err := os.Stat(cacheRoot); os.IsNotExist(err) {
			log.Println("ezgo cache directory does not exist. Nothing to do.")
			return
		}
		log.Printf("Removing ezgo cache directory: %s", cacheRoot)
		if err := os.RemoveAll(cacheRoot); err != nil {
			log.Fatalf("!!! ezgo: Failed to remove cache directory: %v", err)
		}
		log.Println("Cache cleaned successfully.")
	case "path":
		fmt.Println(cacheRoot)
	case "vars":
		cgoEnv, err := SetupCGOEnvironment(isQuiet, nil)
		if err != nil {
			log.Fatalf("!!! ezgo: Failed to setup environment to read variables: %v", err)
		}
		// We call GetGoEnv with a nil config to get the raw CGO variables.
		envVars := cgoEnv.GetGoEnv(nil)
		for _, v := range envVars {
			// Print only the CGO-specific variables for clarity.
			if strings.HasPrefix(v, "CGO_") || strings.HasPrefix(v, "CC=") || strings.HasPrefix(v, "CXX=") {
				fmt.Println(v)
			}
		}

	default:
		log.Fatalf("Unknown command: 'ezgo env %s'. Use 'clean', 'path', or 'vars'.", args[0])
	}
}
