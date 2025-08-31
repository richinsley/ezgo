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

## **Installation**

Ensure you have Go installed and your GOPATH is set up correctly, then simply run:

```
go install github.com/richinsley/ezgo@latest
```

## **Quick Start Guide**

Let's build a simple Go application that uses GLFW to open a window.

**1\. Initialize your project**

Create a new directory for your project and run ezgo mod init. This will create your project's .ezgo.yml file.

mkdir my-glfw-app  
cd my-glfw-app  
ezgo mod init

**2\. Add a C dependency**

We need the glfw library. Add it using ezgo pkg add. ezgo will automatically download and install it into its managed environment.

ezgo pkg add glfw

Your .ezgo.yml will now look like this:

\# Add conda-forge package names for your CGO project  
packages:  
  \- glfw  
\# ...

**3\. Write your Go code**

Create a main.go file with the following code:

```go
package main

/*
#cgo LDFLAGS: -lglfw3 -lgdi32
// The above linker flags are for MinGW on Windows.
*/
import "C"
import "[github.com/go-gl/glfw/v3.3/glfw](https://github.com/go-gl/glfw/v3.3/glfw)"

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

ezgo build \-o myapp.exe .

ezgo will compile your application and automatically copy the required glfw3.dll into the same directory as myapp.exe.

**5\. Run it\!**

Your application is now a portable, self-contained executable.

.\\myapp.exe

## **Command Reference**

### **Core Commands**

* ezgo build: Compiles your Go project. Accepts all flags that go build does.  
  * \-no-copy: An optional flag to prevent the automatic copying of DLLs. Useful for development when you plan to run the app inside an ezgo shell.  
* ezgo run: Compiles and runs your Go program.  
* ezgo test: Runs your project's tests.

### **Module and Package Management**

* ezgo mod init: Creates a .ezgo.yml file in the current directory.  
* ezgo pkg add \<pkg...\>: Adds one or more Conda packages to your .ezgo.yml and installs them.  
* ezgo pkg tidy: Ensures all packages listed in your .ezgo.yml are installed in the environment.

### **Environment and Shell**

* ezgo shell \[powershell|cmd\]: Starts a new, interactive shell (cmd.exe by default) with the CGO environment fully activated. Perfect for debugging or manual compilation.  
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

## **How It Works**

ezgo is powered by [Micromamba](https://mamba.readthedocs.io/en/latest/user_guide/micromamba.html), a fast, native, and self-contained package manager. On its first run, ezgo downloads and sets up a private Micromamba instance and uses it to create a sandboxed environment containing the MinGW GCC toolchain. All subsequent commands use this stable, pre-configured environment to build your projects.

## **License**

This project is licensed under the MIT License.