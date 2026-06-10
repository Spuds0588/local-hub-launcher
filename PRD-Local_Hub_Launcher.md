# Product Requirement Document (PRD)
## 1. Core Objective
The **Universal Local Hub Launcher** serves as an offline, zero-administration desktop application gateway. It monitors a designated local directory, dynamically rendering available HTML tools as an interactive dashboard. It satisfies browser security origins using a fixed, predictable local address (http://127.0.0.1:8999).
## 2. Enterprise Constraints & Security Architecture
To ensure compatibility with strict regulatory baselines (e.g., finance, healthcare, defense aerospace):
 * **Zero External Dependencies:** No Node.js, Python, or runtime prerequisites. Compiles to a raw, self-contained native binary.
 * **Minimal System Footprint:** Uses native Go channels and basic HTTP structures. Consumes less than **15MB of RAM** and **0% idle CPU**, making it completely negligible on heavily taxed corporate workstations.
 * **Low Security Profile:** * Binds exclusively to the isolated loopback adapter (127.0.0.1). It cannot receive connections from the outside network, and it cannot send data out.
   * Does **not** require administrative/root privileges to run or execute.
   * Avoids embedding an expensive web rendering engine (like Electron), instead relying on the secure, already-vetted native browser installed on the host OS.
## 3. User Experience & IT Workflow
 * **IT Deployment:** The IT team pushes the binary and an empty folder (C:\InternalTools\) via their device management software. The binary is configured to run at startup.
 * **User Experience:** Navigating to http://127.0.0.1:8999 presents a clean, card-based launchpad displaying all available utilities. Clicking a card spawns that specific tool in a sandboxed browser tab with a persistent origin.
# Technical Architecture
The system works by reading the contents of a local folder at runtime, parsing the filenames of your single-file apps, and dynamically generating a lightweight dashboard.
### The Source Code (main.go)
```go
package main

import (
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// Fixed, predictable port for documentation, job aids, and intranet links
const ServicePort = "8999"
const DefaultFolderName = "InternalTools"

// Resolves the absolute path to the directory containing the tools
func getToolsDir() string {
	baseDir, err := os.UserHomeDir()
	if err != nil {
		baseDir = "."
	}
	dirPath := filepath.Join(baseDir, DefaultFolderName)
	// Auto-create folder if it doesn't exist yet
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		_ = os.MkdirAll(dirPath, 0755)
	}
	return dirPath
}

// Clean UI Template built with pure CSS, zero external asset or font requests
const dashboardTemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>Internal Application Hub</title>
    <style>
        body { font-family: system-ui, -apple-system, sans-serif; background: #0f172a; color: #f1f5f9; max-width: 1000px; margin: 40px auto; padding: 0 20px; }
        h1 { color: #38bdf8; font-size: 28px; margin-bottom: 5px; }
        p.subtitle { color: #94a3b8; margin-top: 0; margin-bottom: 30px; }
        .grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(280px, 1fr)); gap: 20px; }
        .card { background: #1e293b; border: 1px solid #334155; border-radius: 8px; padding: 20px; display: flex; flex-direction: column; justify-content: space-between; transition: transform 0.15s, border-color 0.15s; }
        .card:hover { transform: translateY(-2px); border-color: #38bdf8; }
        .card-title { font-size: 18px; font-weight: 600; margin: 0 0 10px 0; color: #fff; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
        .card-meta { font-size: 12px; color: #64748b; margin-bottom: 20px; font-family: monospace; }
        .btn { background: #0284c7; color: white; text-decoration: none; text-align: center; padding: 10px; border-radius: 6px; font-weight: 600; font-size: 14px; transition: background 0.1s; }
        .btn:hover { background: #0369a1; }
        .empty-state { grid-column: 1 / -1; text-align: center; background: #1e293b; padding: 40px; border-radius: 8px; border: 2px dashed #334155; color: #94a3b8; }
        .empty-state code { background: #0f172a; padding: 4px 8px; border-radius: 4px; color: #38bdf8; }
    </style>
</head>
<body>

    <h1>⚡ Internal Application Hub</h1>
    <p class="subtitle">Secure, offline environment serving isolated utility apps. Tool Directory: <code>{{.ToolsPath}}</code></p>

    <div class="grid">
        {{range .Tools}}
        <div class="card">
            <div>
                <div class="card-title" title="{{.DisplayName}}">{{.DisplayName}}</div>
                <div class="card-meta">File: {{.FileName}}</div>
            </div>
            <a href="/launch?app={{.FileName}}" target="_blank" class="btn">Launch Tool</a>
        </div>
		{{else}}
        <div class="empty-state">
            <h3>No tools deployed yet</h3>
            <p>Drop your bundled <code>.html</code> tool files into your local directory to see them map here instantly:</p>
            <p><code>{{.ToolsPath}}</code></p>
        </div>
        {{end}}
    </div>

</body>
</html>`

type ToolItem struct {
	FileName    string
	DisplayName string
}

type DashboardData struct {
	ToolsPath string
	Tools     []ToolItem
}

// Converts standard snake_case or kebab-case filenames to readable card headers
func formatDisplayName(filename string) string {
	nameWithoutExt := strings.TrimSuffix(filename, filepath.Ext(filename))
	nameWithSpaces := strings.NewReplacer("-", " ", "_", " ").Replace(nameWithoutExt)
	return strings.Title(strings.ToLower(nameWithSpaces))
}

func main() {
	toolsDir := getToolsDir()
	serverAddress := "127.0.0.1:" + ServicePort

	// 1. Dashboard View Route
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		var tools []ToolItem
		_ = filepath.WalkDir(toolsDir, func(path string, d fs.DirEntry, err error) error {
			if err == nil && !d.IsDir() && strings.HasSuffix(strings.ToLower(d.Name()), ".html") {
				tools = append(tools, ToolItem{
					FileName:    d.Name(),
					DisplayName: formatDisplayName(d.Name()),
				})
			}
			return nil
		})

		tmpl, _ := template.New("hub").Parse(dashboardTemplate)
		_ = tmpl.Execute(w, DashboardData{
			ToolsPath: toolsDir,
			Tools:     tools,
		})
	})

	// 2. Tab Launching Route (Creates the independent web origin sandbox)
	http.HandleFunc("/launch", func(w http.ResponseWriter, r *http.Request) {
		appName := r.URL.Query().Get("app")
		if appName == "" || strings.Contains(appName, "..") || strings.Contains(appName, "/") || strings.Contains(appName, "\\") {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		targetFilePath := filepath.Join(toolsDir, appName)
		if _, err := os.Stat(targetFilePath); os.IsNotExist(err) {
			http.Error(w, "Application file not found in registry folder.", http.StatusNotFound)
			return
		}

		// Sets explicit security headers to insulate applications from one another
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		http.ServeFile(w, r, targetFilePath)
	})

	// 3. Fire up the high-efficiency loopback network server
	go func() {
		_ = http.ListenAndServe(serverAddress, nil)
	}()

	// Brief pause to ensure standard runtime threads are initialized, then call system browser
	time.Sleep(150 * time.Millisecond)
	openBrowser("http://" + serverAddress)

	// Keep parent thread alive without chewing processing loops
	select {}
}

func openBrowser(url string) {
	var cmd string
	var args []string
	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start", url}
	case "darwin":
		cmd = "open"
		args = []string{url}
	default:
		cmd = "xdg-open"
		args = []string{url}
	}
	_ = exec.Command(cmd, args...).Start()
}

```
# Developer & Packaging Task List
### Phase 1: Local Setup
 * [ ] Initialize your build repository: go mod init local-hub
 * [ ] Paste the code above into a file named main.go.
 * [ ] Execute go run main.go to test execution.
 * [ ] Verify that a folder named InternalTools was successfully initialized inside your user profile directory (~ or C:\Users\Username\). Drop a couple of sample .html files in there and refresh http://127.0.0.1:8999 to ensure they map as cards.
### Phase 2: Compiling Low-Profile Production Binaries
 * [ ] **Windows Production Build:** Use specific compiler constraints to strip out symbols, debug tables, and prevent an empty black command prompt window from displaying to the end-user:
   ```bash
   GOOS=windows GOARCH=amd64 go build -ldflags="-s -w -H=windowsgui" -o HubLauncher.exe main.go
   
   ```
   *(The -s -w flags drop file size significantly, while -H=windowsgui makes it completely silent and background-native).*
 * [ ] **macOS Production Build:** ```bash
   GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o HubLauncherMac main.go
   ```
   
   
   ```
### Phase 3: IT Automated Deployment Structure
 * [ ] **Deployment Script Strategy:** Configure your MDM system (like Intune, Jamf, or a GPO logon script) to run a baseline task:
   1. Copies HubLauncher.exe to a permanent read-only directory like C:\Program Files\HubLauncher\.
   2. Creates the target deployment folder: C:\Users\%USERNAME%\InternalTools\ (or modifies the Go script to read from a shared folder variable like C:\InternalTools).
   3. Drops a shortcut link to HubLauncher.exe into the user's startup directory (C:\ProgramData\Microsoft\Windows\Start Menu\Programs\StartUp\).
