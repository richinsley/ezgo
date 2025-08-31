# **EZGO \- The Easy Way to CGO on Windows**

**Tired of fighting with MSYS2, MinGW, and PATH variables just to build a CGO project on Windows? ezgo is your solution.**

ezgo is a zero-configuration, CGO-aware wrapper for the standard Go compiler. It automatically downloads and manages a complete MinGW GCC toolchain and any C/C++ dependencies your project needs, letting you focus on coding instead of configuration.

## **Why ezgo?**

Using C libraries in Go with CGO on Windows is a notoriously painful experience. Developers often struggle with:

* Installing and configuring the correct MinGW toolchain.  
* Managing complex PATH environment variables.  
* Finding and installing C/C++ library dependencies like FFmpeg, GLFW, or ZeroMQ.  
* Ensuring that the final compiled application can find its required .dll files at runtime.

ezgo solves all of these problems by providing a simple, go mod-like workflow for your C dependencies.

## **Features**

* **Automatic Toolchain:** Downloads a self-contained, up-to-date MinGW GCC toolchain on first run. No manual installation needed.  
* **Dependency Management:** Manage C/C++ libraries from the massive Conda-Forge repository using a simple .ezgo.yml file.  
* **Portable Builds:** Automatically finds and copies all required .dll files next to your compiled executable for easy distribution.  
* **Interactive Shell:** Jump into a fully configured shell session where all CGO environment variables are correctly set.  
* **Familiar Commands:** Uses an intuitive command structure like ezgo build, ezgo mod init, and ezgo pkg add.  
* **Complete Development Environment:** The managed environment isn't just for C libraries. It includes a Python interpreter, and you can add other essential tools like cmake or make directly from Conda-Forge.

## **Platform Support**

ezgo is designed exclusively for **Windows on amd64 (x86\_64)** processors.

While Go itself supports Windows on Arm64, ezgo does not natively support this platform. This is because the Conda-Forge repository, which ezgo relies on for the C/C++ toolchain and libraries, has limited package availability for win-arm64.

However, you can still use ezgo on Windows on Arm devices (like the Surface Pro X) through the built-in x64 emulation layer. The tool will function correctly, but performance may be impacted by the translation layer.

## **Installation**

Ensure you have Go installed and your GOPATH is set up correctly, then simply run:
```
go install github.com/richinsley/ezgo@latest
```
## **Quick Start Guide**

Let's build a simple Go application that uses GLFW to open a window.

**1\. Initialize your project**

Create a new directory for your project and run ezgo mod init. This will create your project's .ezgo.yml file.
```
mkdir my-glfw-app  
cd my-glfw-app  
ezgo mod init
```
**2\. Add a C dependency**

We need the glfw library. Add it using ezgo pkg add. ezgo will automatically download and install it into its managed environment.
```
ezgo pkg add glfw
```
Your .ezgo.yml will now look like this:
```yaml
# Add conda-forge package names for your CGO project  
packages:  
  - glfw  
# ...
```
**3\. Write your Go code**

Create a main.go file with the following code:
```go
package main

/*
#cgo LDFLAGS: -lglfw3 -lgdi32
// The above linker flags are for MinGW on Windows.
*/
import "C"
import "github.com/go-gl/glfw/v3.3/glfw"

func main() {
  if err := glfw.Init(); err != nil {
    panic(err)
  }
  defer glfw.Terminate()

  window, err := glfw.CreateWindow(640, 480, "Hello from EZGO!", nil, nil)  
  if err != nil {  
      panic(err)  
  }

  window.MakeContextCurrent()

  for window.ShouldClose() {  
      glfw.PollEvents()  
  }  
}
```

*(Note: You'll also need to run go get github.com/go-gl/glfw/v3.3/glfw for the Go bindings.)*

**4\. Build your application**

Use ezgo build just like you would use go build.
```
ezgo build -o myapp.exe .
```
ezgo will compile your application and automatically copy the required glfw3.dll into the same directory as myapp.exe.

**5\. Run it\!**

Your application is now a portable, self-contained executable.
```
.\myapp.exe
```
## **Development Workflow: The ezgo shell**

For day-to-day development, constantly copying DLLs after every build is slow. The recommended workflow is to use the ezgo shell, which creates a special command prompt where all your C libraries "just work."

**1\. Build Without Copying DLLs**

Use the \-no-copy flag to compile your app quickly. This creates the .exe file but skips finding and copying all the dependency .dll files.
```
ezgo build -no-copy -o myapp.exe .
```
Your project directory stays clean, containing only your source code and the single executable.

**2\. Start the EZGO Shell**

Next, start the pre-configured shell.
```bash
ezgo shell
# for powershell
ezgo shell powershell
```
This opens a new command prompt that looks normal, but behind the scenes, ezgo has temporarily updated the PATH to point to the folders containing all the necessary C library DLLs (like glfw3.dll).

**3\. Run Your App (or Tests)**

Now, inside this special shell, you can run your executable directly. Windows will automatically find the required DLLs because of the updated PATH.
```
.\myapp.exe
```

This workflow is also perfect for running tests or using debuggers that need to find the C libraries at runtime.

When you are finished and ready to share your application, run a final ezgo build without the \-no-copy flag to create a portable folder with your executable and all its required DLLs.

## **Command Reference**

### **Core Commands**

* ezgo build: Compiles your Go project. Accepts all flags that go build does.  
  * \-no-copy: **(Recommended for development)** Prevents the automatic copying of DLLs. Use this for faster iteration, then run your app inside an ezgo shell session.  
* ezgo run: Compiles and runs your Go program. Because it runs in a configured environment, DLLs are found automatically.  
* ezgo test: Runs your project's tests in a configured environment, ensuring any C libraries are found.

### **Module and Package Management**

* ezgo mod init: Creates a .ezgo.yml file in the current directory.  
* ezgo pkg add \<pkg...\>: Adds one or more Conda packages to your .ezgo.yml and installs them.  
* ezgo pkg tidy: Ensures all packages listed in your .ezgo.yml are installed in the environment.

### **Environment and Shell**

* ezgo shell \[powershell|cmd\]: Starts a a new, interactive shell (cmd.exe by default) with the CGO environment fully activated. This is the **cornerstone of the development workflow**, allowing you to run executables compiled with \-no-copy and use tools like go test or debuggers.  
* ezgo env clean: Deletes the entire cached ezgo environment from your user profile.  
* ezgo env path: Prints the root path of the ezgo cache.  
* ezgo env vars: Prints the CGO-specific environment variables (CC, CXX, CGO\_CFLAGS, etc.) that ezgo uses.

## **The .ezgo.yml File**

This file controls your project's C/C++ dependencies and build settings.
```yaml
# A list of conda-forge package names for your CGO project.  
# Find packages at anaconda.org  
packages:  
  - ffmpeg  
  - zeromq

# Custom environment variables to be passed to the go compiler.  
# These will be available during the 'go build' process.  
environment:  
  SOME_CUSTOM_FLAG: "true"
```
## **Beyond C Libraries: Managing Build Tools**

The ezgo environment is more than just a home for your C/C++ libraries. Because it's powered by Micromamba, you get a full-featured development environment out of the box.

**Python Included**

Every ezgo environment automatically includes a Python interpreter. You can use it for scripting or running build tools directly inside the ezgo shell.
```
ezgo shell  
(cgo_win_env_py312) C:\Users\YourUser\my-project\> python --version  
Python 3.12.x
```
**Adding Tools like CMake**

If your project requires other build tools like cmake, you don't need to install them system-wide. Just add them to your .ezgo.yml file:
```yaml
packages:  
  - glfw  
  - cmake  # Add cmake to your project
```
Then, run ezgo pkg tidy to install it. Now, cmake is available inside your project's shell:
```
ezgo pkg tidy  
ezgo shell  
(cgo_win_env_py312) C:\Users\YourUser\my-project\> cmake --version  
cmake version 3.xx.x
```
## **Troubleshooting**

### **Aggressive Antivirus / Windows Defender Issues**

It is a very common and frustrating issue for Go developers on Windows that antivirus software, particularly Windows Defender, can incorrectly flag Go programs as malicious. This can manifest in two ways:

1. **Builds Fail:** You will see an error mentioning a "virus or potentially unwanted software" when the Go compiler tries to create an executable.  
2. **Tools Disappear:** The antivirus may silently quarantine or delete the tools themselves, including ezgo.exe or even go.exe. If the ezgo command suddenly stops working, this is the likely cause.

This happens because Go compiles new, unsigned executables, and aggressive "heuristic" or "cloud-based" protection flags this behavior as suspicious.

**The Complete Solution:** The most reliable way to permanently solve this is to add exclusions for all the locations Go uses for its tools and build processes.

#### **PowerShell Method (Recommended)**

1. Open PowerShell **as an Administrator**.  
2. Run the following commands. They will automatically find your Go paths and add the necessary exclusions.
```powershell
   # 1. Exclude the Go installation folder itself (GOROOT)  
   Add-MpPreference -ExclusionPath (Get-Command go).Source.Substring(0, (Get-Command go).Source.LastIndexOf('\'))

   # 2. Exclude your GOPATH, where your source code and tools are installed  
   Add-MpPreference -ExclusionPath (go env GOPATH)

   # 3. Exclude the Go build cache  
   Add-MpPreference -ExclusionPath "$env:LOCALAPPDATA\go-build"

   # 4. Exclude the system temporary folder  
   Add-MpPreference -ExclusionPath "$env:TEMP"
```
#### **Reinstalling ezgo**

If ezgo was quarantined by your antivirus, you will need to reinstall it after adding the exclusions:
```
go install github.com/richinsley/ezgo@latest
```
#### **GUI Method**

1. Open **Windows Security** and go to **Virus & threat protection**.  
2. Under **Virus & threat protection settings**, click **Manage settings**.  
3. Scroll down to **Exclusions** and click **Add or remove exclusions**.  
4. Click **Add an exclusion** \> **Folder** and add each of the following paths. You can find the exact paths by running go env GOROOT and go env GOPATH in your terminal.  
   * Your GOROOT folder (e.g., C:\\Program Files\\Go)  
   * Your GOPATH folder (e.g., C:\\Users\\YourUser\\go)  
   * %LOCALAPPDATA%\\go-build  
   * %TEMP%

This comprehensive set of exclusions should prevent any further interference from your antivirus.

## **How It Works**

ezgo is powered by [Micromamba](https://mamba.readthedocs.io/en/latest/user_guide/micromamba.html), a fast, native, and self-contained package manager. On its first run, ezgo downloads and sets up a private Micromamba instance and uses it to create a sandboxed environment containing the MinGW GCC toolchain. All subsequent commands use this stable, pre-configured environment to build your projects.

## **License**

This project is licensed under the MIT License.