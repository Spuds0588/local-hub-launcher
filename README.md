# Universal Local Hub Launcher

<p align="center">
  <img src="assets/logo.png" width="120" height="120" alt="Universal Local Hub Launcher Logo">
</p>

A lightweight, zero-administration desktop application gateway for developer utility tools. It monitors a local directory (`~/InternalTools`) and dynamically maps available single-file HTML apps onto an interactive dashboard served locally on port `8999`.

Built with high-compliance and corporate governance in mind, it operates entirely offline with **zero system runtimes**, binds to the isolated loopback interface, and isolates applications via local browser sandbox controls.

---

## ⚡ Key Features

* **Zero Administration:** Compiles to a single self-contained native binary. No Node.js, Python, or runtime configuration prerequisites required.
* **Ultra Lightweight:** Uses native Go channels. Consumes **<15MB of RAM** and **0% idle CPU**, leaving corporate workstations completely untaxed.
* **Isolated Origin Security:** Bypasses `file://` cookie/local storage leakage. Serves local tools under isolated paths with strict `X-Content-Type-Options: nosniff` and `X-Frame-Options: DENY` headers.
* **Strict Loopback Binding:** Binds strictly to the loopback adapter (`127.0.0.1`). It cannot receive connections from outside networks, nor can it transmit data.
* **Direct File Mapping:** Automatically renders dashboard cards from single-file HTML tools (`.html`) dropped in the monitored workspace folder.

---

## 🚀 Quick Start

1. **Monitored Directory:**
   By default, the launcher maps files from `~/InternalTools` (e.g., `/home/username/InternalTools` or `C:\Users\username\InternalTools`). The application will automatically initialize this directory on start if it does not exist.

2. **Drop HTML Tools:**
   Drop your bundled `.html` tools (e.g., JSON formatters, database viewers, REPLs) into the `InternalTools` folder.

3. **Run the Application:**
   If you have Go installed, execute:
   ```bash
   go run main.go
   ```
   Or execute the compiled binary. The launcher automatically opens your default system browser and redirects to the landing dashboard:
   ```
   http://127.0.0.1:8999/
   ```

---

## 🛠️ Cross-Compilation & Packaging

Compile optimized, stripped binaries for your target deployment architectures:

### Windows (Silent Background Launch)
To prevent a command prompt window from flashing on the user's screen, use the `-H=windowsgui` flag:
```bash
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w -H=windowsgui" -o HubLauncher.exe main.go
```
*The `-s -w` flags strip debugging information to reduce the binary size to ~1-2 MB.*

### macOS (Darwin)
```bash
GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o HubLauncherMac main.go
```

### Linux
```bash
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o HubLauncherLinux main.go
```

---

## 🏢 IT Automated Deployment

For organization-wide rollouts using systems management platforms (such as Microsoft Intune, GPO, Jamf, or PDQ):

1. **Distribute Binary:** Push the compiled binary (e.g. `HubLauncher.exe`) to a read-only local program directory:
   `C:\Program Files\HubLauncher\HubLauncher.exe`
2. **Setup Folder:** Create the user tool directory:
   `C:\Users\%USERNAME%\InternalTools\`
3. **Configure Startup:** Place a shortcut (`.lnk`) pointing to the executable inside the Windows Startup folder to ensure it runs silently on login:
   `C:\ProgramData\Microsoft\Windows\Start Menu\Programs\StartUp\`
4. **Deploy Tools:** Push single-file tools (like `calculator.html`, `json-viewer.html`) into the user's `InternalTools` directory. The launcher will automatically discover them at runtime.

---

## 📄 License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.
