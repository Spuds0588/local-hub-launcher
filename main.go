package main

import (
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
