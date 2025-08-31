package pkg

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	jumpboot "github.com/richinsley/jumpboot"
	"gopkg.in/yaml.v3"
)

// ProjectConfig defines the structure of the .ezgo.yml file.
type ProjectConfig struct {
	Packages    []string          `yaml:"packages,omitempty"`
	Environment map[string]string `yaml:"environment,omitempty"`
}

// CGOEnvironment holds the necessary paths and settings for a CGO build environment.
type CGOEnvironment struct {
	*jumpboot.Environment
	CompilerBinPath string
	ToolBinPath     string
	FullBinPath     string
	IncludePath     string
	LibPath         string
}

func LoadProjectConfig() (*ProjectConfig, error) {
	configFile := ".ezgo.yml"
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		return nil, nil // No config file is not an error
	}
	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, err
	}
	var cfg ProjectConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("could not parse .ezgo.yml: %w", err)
	}
	return &cfg, nil
}

// A set of common Windows system DLLs that should not be copied.
var systemDLLs = map[string]bool{
	"advapi32.dll": true, "comdlg32.dll": true, "gdi32.dll": true,
	"kernel32.dll": true, "msvcrt.dll": true, "ole32.dll": true,
	"oleaut32.dll": true, "shell32.dll": true, "user32.dll": true,
	"winmm.dll": true, "ws2_32.dll": true, "ntdll.dll": true,
	"rpcrt4.dll": true, "shlwapi.dll": true,
}

// getDependencies runs objdump on a binary and returns its non-system DLL dependencies.
func GetDependencies(binaryPath, objdumpPath string) (map[string]bool, error) {
	deps := make(map[string]bool)
	cmd := exec.Command(objdumpPath, "-p", binaryPath)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to run objdump on %s: %w", binaryPath, err)
	}

	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "DLL Name:") {
			dllName := strings.TrimSpace(strings.TrimPrefix(line, "DLL Name:"))
			if !systemDLLs[strings.ToLower(dllName)] {
				deps[dllName] = true
			}
		}
	}
	return deps, nil
}

// HandlePostBuild recursively finds and copies all required runtime DLLs.
func HandlePostBuild(args []string, cgoEnv *CGOEnvironment, isQuiet bool, skipCopy bool) error {
	if skipCopy {
		if !isQuiet {
			log.Println("--- ezgo: Build successful. Skipping DLL copy due to -no-copy flag.")
		}
		return nil
	}

	var outputPath string
	for i, arg := range args {
		if arg == "-o" && i+1 < len(args) {
			outputPath = args[i+1]
			break
		}
	}
	if outputPath == "" {
		cwd, _ := os.Getwd()
		outputPath = filepath.Join(cwd, filepath.Base(cwd)+".exe")
	}
	if !filepath.IsAbs(outputPath) {
		cwd, _ := os.Getwd()
		outputPath = filepath.Join(cwd, outputPath)
	}

	destDir := filepath.Dir(outputPath)
	objdumpPath := filepath.Join(cgoEnv.CompilerBinPath, "objdump.exe")

	// --- Recursive Dependency Resolution ---
	dllsToProcess := []string{}
	allRequiredDLLs := make(map[string]bool)
	processedDLLs := make(map[string]bool)

	// Get initial dependencies from the main executable.
	initialDeps, err := GetDependencies(outputPath, objdumpPath)
	if err != nil {
		return err
	}
	for dll := range initialDeps {
		if !allRequiredDLLs[dll] {
			allRequiredDLLs[dll] = true
			dllsToProcess = append(dllsToProcess, dll)
		}
	}

	searchPaths := []string{
		filepath.Join(cgoEnv.EnvPath, "Library", "bin"),
		cgoEnv.ToolBinPath,
	}

	// Process the queue of DLLs to find their dependencies.
	for len(dllsToProcess) > 0 {
		dll := dllsToProcess[0]
		dllsToProcess = dllsToProcess[1:]

		if processedDLLs[dll] {
			continue
		}
		processedDLLs[dll] = true

		var dllPath string
		for _, p := range searchPaths {
			path := filepath.Join(p, dll)
			if _, err := os.Stat(path); err == nil {
				dllPath = path
				break
			}
		}

		if dllPath == "" {
			continue // Could be a system DLL or already handled, skip.
		}

		// Get dependencies of the current DLL.
		transitiveDeps, err := GetDependencies(dllPath, objdumpPath)
		if err != nil {
			if !isQuiet {
				log.Printf("--- ezgo: warning: could not analyze dependencies for %s: %v", dll, err)
			}
			continue
		}

		// Add new, unseen dependencies to the queue.
		for newDep := range transitiveDeps {
			if !allRequiredDLLs[newDep] {
				allRequiredDLLs[newDep] = true
				dllsToProcess = append(dllsToProcess, newDep)
			}
		}
	}

	if !isQuiet {
		log.Printf("--- ezgo: Found %d required DLL(s). Copying to %s", len(allRequiredDLLs), destDir)
	}

	// Copy the final, complete set of DLLs.
	copiedCount := 0
	for dll := range allRequiredDLLs {
		found := false
		for _, searchPath := range searchPaths {
			srcFile := filepath.Join(searchPath, dll)
			if _, err := os.Stat(srcFile); err == nil {
				destFile := filepath.Join(destDir, dll)
				if err := copyFile(srcFile, destFile); err != nil {
					return fmt.Errorf("failed to copy DLL %s: %w", dll, err)
				}
				copiedCount++
				found = true
				break
			}
		}
		if !found && !isQuiet {
			log.Printf("--- ezgo: warning: could not find required DLL %s", dll)
		}
	}

	if !isQuiet && copiedCount > 0 {
		log.Printf("--- ezgo: Copied %d DLLs.", copiedCount)
	}
	return nil
}

// copyFile is a simple helper to copy a file from source to destination.
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}
