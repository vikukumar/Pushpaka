package worker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	"github.com/vikukumar/Pushpaka/pkg/basemodel"
	"github.com/vikukumar/Pushpaka/pkg/models"
	"github.com/vikukumar/Pushpaka/worker/internal/config"
)

// JobReporter is called on job lifecycle events (may be nil).
type JobReporter interface {
	JobStarted()
	JobFinished()
}

// processManager tracks running direct-deployment processes (no Docker).
var processManager = struct {
	sync.Mutex
	procs map[string]*os.Process // deploymentID -> running process
}{procs: make(map[string]*os.Process)}

const deployJobQueue = "pushpaka:deploy:queue"

type BuildWorker struct {
	id              int
	db              *gorm.DB
	rdb             *redis.Client
	cfg             *config.Config
	dockerAvailable bool
	reporter        JobReporter
}

func NewBuildWorker(id int, db *gorm.DB, rdb *redis.Client, cfg *config.Config) *BuildWorker {
	return &BuildWorker{
		id:              id,
		db:              db,
		rdb:             rdb,
		cfg:             cfg,
		dockerAvailable: checkDockerAvailable(cfg.DockerHost),
	}
}

// checkDockerAvailable tries to connect to the Docker daemon.
// On Linux/Mac it checks the socket; on Windows the named pipe.
// Falls back to running `docker info` as a last resort.
func checkDockerAvailable(dockerHost string) bool {
	// Prefer direct socket/pipe check (no subprocess overhead).
	socketPath := "/var/run/docker.sock"
	if runtime.GOOS == "windows" {
		socketPath = `\\.\pipe\docker_engine`
	}
	if dockerHost != "" {
		// Strip scheme prefix e.g. "unix:///var/run/docker.sock" -> "/var/run/docker.sock"
		h := strings.TrimPrefix(dockerHost, "unix://")
		h = strings.TrimPrefix(h, "npipe://")
		socketPath = h
	}

	var connectable bool
	if runtime.GOOS == "windows" {
		// Named pipe  just try to dial
		conn, err := net.DialTimeout("unix", socketPath, 2*time.Second)
		if err == nil {
			conn.Close()
			connectable = true
		}
	} else {
		if _, err := os.Stat(socketPath); err == nil {
			conn, err := net.DialTimeout("unix", socketPath, 2*time.Second)
			if err == nil {
				conn.Close()
				connectable = true
			}
		}
	}
	if connectable {
		return true
	}

	// Fallback: try `docker info` CLI
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "docker", "info")
	if dockerHost != "" {
		cmd.Env = append(os.Environ(), "DOCKER_HOST="+dockerHost)
	}
	return cmd.Run() == nil
}

// DockerAvailable reports whether Docker was detected at worker startup.
func (w *BuildWorker) DockerAvailable() bool { return w.dockerAvailable }

func (w *BuildWorker) Run(ctx context.Context) {
	log.Info().Int("worker_id", w.id).Msg("build worker started")
	for {
		select {
		case <-ctx.Done():
			log.Info().Int("worker_id", w.id).Msg("build worker stopping")
			return
		default:
			// Blocking pop from Redis queue with 5s timeout
			result, err := w.rdb.BRPop(ctx, 5*time.Second, deployJobQueue).Result()
			if err != nil {
				if err == redis.Nil {
					continue // timeout, try again
				}
				if ctx.Err() != nil {
					return // context cancelled
				}
				log.Error().Err(err).Msg("redis brpop error")
				continue
			}

			if len(result) < 2 {
				continue
			}

			var job models.DeploymentJob
			if err := json.Unmarshal([]byte(result[1]), &job); err != nil {
				log.Error().Err(err).Msg("failed to unmarshal job")
				continue
			}

			w.processJob(ctx, &job)
		}
	}
}

// RunInProcess reads jobs from an in-process channel instead of Redis.
// reporter is optional (may be nil); when non-nil its JobStarted/JobFinished
// methods are called around each processed job.
func (w *BuildWorker) RunInProcess(ctx context.Context, ch <-chan []byte, reporter JobReporter) {
	w.reporter = reporter
	log.Info().
		Int("worker_id", w.id).
		Bool("docker", w.dockerAvailable).
		Msg("in-process build worker started")
	for {
		select {
		case <-ctx.Done():
			log.Info().Int("worker_id", w.id).Msg("in-process build worker stopping")
			return
		case payload, ok := <-ch:
			if !ok {
				return
			}
			var job models.DeploymentJob
			if err := json.Unmarshal(payload, &job); err != nil {
				log.Error().Err(err).Msg("failed to unmarshal in-process job")
				continue
			}
			if reporter != nil {
				reporter.JobStarted()
			}
			w.processJob(ctx, &job)
			if reporter != nil {
				reporter.JobFinished()
			}
		}
	}
}

func (w *BuildWorker) processJob(ctx context.Context, job *models.DeploymentJob) {
	logger := log.With().
		Str("deployment_id", job.DeploymentID).
		Str("project_id", job.ProjectID).
		Int("worker_id", w.id).
		Logger()

	logger.Info().Bool("docker", w.dockerAvailable).Msg("starting build")

	// Update status to building
	w.updateStatus(job.DeploymentID, "building", "")
	w.appendLog(job.DeploymentID, "info", "system", "Build started")

	if !w.dockerAvailable {
		w.appendLog(job.DeploymentID, "info", "system", "Docker not available -- deploying directly (no containerization)")
	}

	// Work directory -- remove any leftover from a failed previous attempt,
	// then let git clone create it fresh.
	workDir := filepath.Join(w.cfg.CloneDir, job.DeploymentID)
	if err := os.RemoveAll(workDir); err != nil {
		w.fail(job.DeploymentID, fmt.Sprintf("failed to clean work dir: %v", err))
		return
	}
	if err := os.MkdirAll(filepath.Dir(workDir), 0755); err != nil {
		w.fail(job.DeploymentID, fmt.Sprintf("failed to create build dir: %v", err))
		return
	}
	// Step 1: Clone repository
	needsClone := !job.IsRecovery
	if job.IsRecovery {
		// Verify if we can actually recover.
		// For Docker: check if image exists. For Direct: check if permanentDir exists.
		canRecover := false
		if w.dockerAvailable {
			checkCmd := exec.CommandContext(ctx, "docker", "inspect", job.ImageTag)
			if err := checkCmd.Run(); err == nil {
				canRecover = true
			}
		} else {
			permanentDir := filepath.Join(w.cfg.DeployDir, job.ProjectID[:8])
			if _, err := os.Stat(permanentDir); err == nil {
				canRecover = true
			}
		}

		if !canRecover {
			w.appendLog(job.DeploymentID, "warn", "system", "Recovery assets not found (image or directory) -- falling back to full build")
			needsClone = true
			// We MUST clear IsRecovery for sub-methods (deployContainer/deployDirect)
			// so they don't try to "skip" steps again.
			job.IsRecovery = false
		} else {
			w.appendLog(job.DeploymentID, "info", "system", "Recovery mode: skipping repository clone")
		}
	}

	if needsClone {
		w.appendLog(job.DeploymentID, "info", "system", fmt.Sprintf("Cloning %s@%s", job.RepoURL, job.Branch))
		if err := w.cloneRepo(ctx, job, workDir); err != nil {
			os.RemoveAll(workDir)
			w.fail(job.DeploymentID, fmt.Sprintf("clone failed: %v", err))
			return
		}
		w.appendLog(job.DeploymentID, "info", "system", "Repository cloned successfully")
	}

	// Capture commit info (only if not recovery)
	if !job.IsRecovery {
		if sha, msg, err := getRepoCommitInfo(workDir); err == nil && sha != "" {
			job.CommitSHA = sha
			w.db.Model(&models.Deployment{}).
				Where("id = ?", job.DeploymentID).
				Updates(map[string]interface{}{
					"commit_sha": sha + " " + msg,
				})
		}
	}

	var containerID, deployURL string
	var deployErr error

	if w.dockerAvailable {
		// Docker path: generate Dockerfile -> build image -> run container
		dockerfilePath := filepath.Join(workDir, "Dockerfile")
		if _, err := os.Stat(dockerfilePath); os.IsNotExist(err) {
			w.appendLog(job.DeploymentID, "info", "system", "No Dockerfile found, generating one...")
			if err := w.generateDockerfile(workDir, job); err != nil {
				os.RemoveAll(workDir)
				w.fail(job.DeploymentID, fmt.Sprintf("dockerfile generation failed: %v", err))
				return
			}
		}

		if job.IsRecovery {
			w.appendLog(job.DeploymentID, "info", "system", "Recovery mode: skipping image build")
		} else {
			w.appendLog(job.DeploymentID, "info", "system", fmt.Sprintf("Building Docker image: %s", job.ImageTag))
			if err := w.buildImage(ctx, job, workDir); err != nil {
				os.RemoveAll(workDir)
				w.fail(job.DeploymentID, fmt.Sprintf("build failed: %v", err))
				return
			}
			w.appendLog(job.DeploymentID, "info", "system", "Docker image built successfully")
		}

		w.appendLog(job.DeploymentID, "info", "system", "Deploying container...")
		containerID, deployURL, deployErr = w.deployContainer(ctx, job)
		os.RemoveAll(workDir)
	} else {
		// Direct path: install deps, build, run process in-place
		w.appendLog(job.DeploymentID, "info", "system", "Installing dependencies and building...")
		containerID, deployURL, deployErr = w.deployDirect(ctx, job, workDir)
		// workDir is kept alive (moved to running dir inside deployDirect)
		os.RemoveAll(workDir)
	}

	if deployErr != nil {
		msg := fmt.Sprintf("deployment failed: %v", deployErr)
		w.fail(job.DeploymentID, msg)
		w.fireNotification(job, "failed", "", msg)
		return
	}

	w.appendLog(job.DeploymentID, "info", "system", fmt.Sprintf("Deployment ID: %s", containerID))
	w.appendLog(job.DeploymentID, "info", "system", fmt.Sprintf("Available at: %s", deployURL))

	// Update deployment as running.
	// For direct (no-Docker) deployments the URL was already set to the proxy
	// path (/app/<projectID>) by the API when the deployment was created, so we
	// must NOT overwrite it here.  We only store external_port so the proxy
	// handler can forward traffic to the right local port.
	//
	// For Docker deployments we update the URL too because Traefik routes it
	// to a Traefik path (/p/<projectID>), different from the initial proxy URL.
	now := time.Now().UTC()
	var dbErr error
	if w.dockerAvailable {
		// Docker: update container_id + URL (Traefik path)
		dbErr = w.db.Model(&models.Deployment{}).
			Where("id = ?", job.DeploymentID).
			Updates(map[string]interface{}{
				"status":       models.DeploymentRunning,
				"container_id": containerID,
				"url":          deployURL,
				"finished_at":  now,
			}).Error
	} else {
		// Direct: keep existing proxy URL, just record container_id + external_port
		extPort := job.Port
		if extPort == 0 {
			extPort = 3000
		}
		dbErr = w.db.Model(&models.Deployment{}).
			Where("id = ?", job.DeploymentID).
			Updates(map[string]interface{}{
				"status":        models.DeploymentRunning,
				"container_id":  containerID,
				"external_port": extPort,
				"finished_at":   now,
			}).Error
	}
	if dbErr != nil {
		logger.Error().Err(dbErr).Msg("failed to update deployment record")
	}

	// Fire notification callback (non-blocking)
	w.fireNotification(job, "running", deployURL, "")

	logger.Info().Str("url", deployURL).Msg("deployment completed")
}

// deployDirect runs the application in-process without Docker.
// It copies the built source to a permanent directory, installs deps, builds,
// starts the process and stores the OS process handle for later cleanup.
func (w *BuildWorker) deployDirect(ctx context.Context, job *models.DeploymentJob, workDir string) (string, string, error) {
	// For zero-downtime, we don't kill the old process yet.
	// The new process will start on a separate port (job.ExternalPort).

	// Determine framework and commands
	buildCmd := job.BuildCommand
	startCmd := job.StartCommand
	port := job.ExternalPort
	if port == 0 {
		port = job.Port
		if port == 0 {
			port = 3000
		}
	}

	pm := ""
	isNodeProject := false
	isPythonProject := false
	pythonReqFile := "" // which Python dependency file was found
	pythonExe := findPythonExe()
	pythonBin := pythonExe // updated to venv python path after install
	var venvBinDir string  // set when venv is created; prepended to PATH for process

	hasFile := func(name string) bool {
		_, err := os.Stat(filepath.Join(workDir, name))
		return err == nil
	}

	// ─── Language / runtime detection ────────────────────────────────────────

	// Mixed-project guard: if the repo contains both a package.json (likely
	// frontend tooling) AND a Python entry file + a Python dependency file,
	// treat the project as Python-primary and skip the Node.js branch.
	isPrimaryPython := false
	if hasFile("package.json") {
		hasPyDep := hasFile("requirements.txt") || hasFile("pyproject.toml") || hasFile("Pipfile")
		if hasPyDep {
			for _, pyEntry := range []string{"main.py", "app.py", "server.py", "manage.py", "asgi.py"} {
				if hasFile(pyEntry) {
					isPrimaryPython = true
					break
				}
			}
		}
	}

	switch {
	case hasFile("package.json") && !isPrimaryPython:
		isNodeProject = true
		pm = detectPackageManager(workDir)
		// Verify the chosen PM binary exists in PATH; fall back to npm.
		if pm != "npm" {
			if _, err := exec.LookPath(pm); err != nil {
				w.appendLog(job.DeploymentID, "warn", "system",
					fmt.Sprintf("'%s' not found in PATH -- falling back to npm", pm))
				pm = "npm"
			}
		}
		// Next.js 15/16: create a minimal config if none exists.
		// Next.js prefers next.config.ts when tsconfig.json is present;
		// otherwise next.config.js (or .mjs for ESM packages).
		hasNextDep := false
		if pkgData, err := os.ReadFile(filepath.Join(workDir, "package.json")); err == nil {
			hasNextDep = strings.Contains(string(pkgData), `"next"`)
		}
		if hasNextDep {
			configExists := false
			for _, name := range []string{"next.config.ts", "next.config.js", "next.config.mjs", "next.config.cjs"} {
				if hasFile(name) {
					configExists = true
					break
				}
			}
			if !configExists {
				if hasFile("tsconfig.json") {
					// TypeScript project: Next.js 15+ tries .ts first.
					tsContent := "import type { NextConfig } from 'next'\nconst nextConfig: NextConfig = {}\nexport default nextConfig\n"
					_ = os.WriteFile(filepath.Join(workDir, "next.config.ts"), []byte(tsContent), 0644)
					w.appendLog(job.DeploymentID, "info", "system", "Created minimal next.config.ts (TypeScript project, no config found)")
				} else {
					// Check for ESM package (\"type\": \"module\")
					isESM := false
					if pkgData, err := os.ReadFile(filepath.Join(workDir, "package.json")); err == nil {
						s := string(pkgData)
						isESM = strings.Contains(s, `"type": "module"`) || strings.Contains(s, `"type":"module"`)
					}
					if isESM {
						esmContent := "/** @type {import('next').NextConfig} */\nconst nextConfig = {}\nexport default nextConfig\n"
						_ = os.WriteFile(filepath.Join(workDir, "next.config.mjs"), []byte(esmContent), 0644)
						w.appendLog(job.DeploymentID, "info", "system", "Created minimal next.config.mjs (ESM project, no config found)")
					} else {
						jsContent := "/** @type {import('next').NextConfig} */\nconst nextConfig = {}\nmodule.exports = nextConfig\n"
						_ = os.WriteFile(filepath.Join(workDir, "next.config.js"), []byte(jsContent), 0644)
						w.appendLog(job.DeploymentID, "info", "system", "Created minimal next.config.js (no config found in repo)")
					}
				}
			}
		}
		if buildCmd == "" {
			buildCmd = pm + " run build"
		}
		if startCmd == "" {
			// Read package.json to determine the best start command.
			startCmd = detectNodeStartCmd(workDir, pm, port)
		}

	case hasFile("requirements.txt") || isPrimaryPython:
		isPythonProject = true
		pythonReqFile = "requirements.txt"
		if startCmd == "" {
			// Check if FastAPI/uvicorn is listed as a dependency.
			if reqBytes, err := os.ReadFile(filepath.Join(workDir, "requirements.txt")); err == nil {
				reqLower := strings.ToLower(string(reqBytes))
				if strings.Contains(reqLower, "fastapi") || strings.Contains(reqLower, "uvicorn") ||
					strings.Contains(reqLower, "starlette") {
					startCmd = fmt.Sprintf("uvicorn %s --host 0.0.0.0 --port %d",
						detectUvicornModule(workDir), port)
				}
			}
			if startCmd == "" {
				startCmd = pythonExe + " " + detectPythonEntry(workDir, port)
			}
		}

	case hasFile("pyproject.toml"):
		isPythonProject = true
		pythonReqFile = "pyproject.toml"
		if startCmd == "" {
			// Check if this is a FastAPI/uvicorn project.
			if content, err := os.ReadFile(filepath.Join(workDir, "pyproject.toml")); err == nil {
				s := string(content)
				if strings.Contains(s, "fastapi") || strings.Contains(s, "uvicorn") {
					startCmd = fmt.Sprintf("uvicorn main:app --host 0.0.0.0 --port %d", port)
				}
			}
			if startCmd == "" {
				startCmd = pythonExe + " " + detectPythonEntry(workDir, port)
			}
		}

	case hasFile("Pipfile"):
		isPythonProject = true
		pythonReqFile = "Pipfile"
		if startCmd == "" {
			startCmd = pythonExe + " " + detectPythonEntry(workDir, port)
		}

	case hasFile("setup.py"):
		isPythonProject = true
		pythonReqFile = "setup.py"
		if startCmd == "" {
			startCmd = pythonExe + " " + detectPythonEntry(workDir, port)
		}

	case hasFile("Cargo.toml"):
		if buildCmd == "" {
			buildCmd = "cargo build --release"
		}
		if startCmd == "" {
			startCmd = "./target/release/app"
		}

	case hasFile("go.mod"):
		if buildCmd == "" {
			buildCmd = "go build -o app ."
		}
		if startCmd == "" {
			if runtime.GOOS == "windows" {
				startCmd = `app.exe`
			} else {
				startCmd = "./app"
			}
		}

	case hasFile("pom.xml"):
		// Java — Maven
		if buildCmd == "" {
			buildCmd = "mvn package -DskipTests -q"
		}
		if startCmd == "" {
			startCmd = fmt.Sprintf("java -jar target/*.jar --server.port=%d", port)
		}

	case hasFile("build.gradle") || hasFile("gradlew"):
		// Java — Gradle
		gradleExe := "gradle"
		if hasFile("gradlew") {
			if runtime.GOOS == "windows" {
				gradleExe = `gradlew.bat`
			} else {
				gradleExe = "./gradlew"
				_ = os.Chmod(filepath.Join(workDir, "gradlew"), 0755)
			}
		}
		if buildCmd == "" {
			buildCmd = gradleExe + " build -x test"
		}
		if startCmd == "" {
			startCmd = fmt.Sprintf("java -jar build/libs/*.jar --server.port=%d", port)
		}

	case hasFile("composer.json"):
		// PHP
		if buildCmd == "" {
			if hasFile("artisan") {
				buildCmd = "composer install --no-dev --optimize-autoloader"
			} else {
				buildCmd = "composer install"
			}
		}
		if startCmd == "" {
			if hasFile("artisan") {
				// Laravel
				startCmd = fmt.Sprintf("php artisan serve --host=0.0.0.0 --port=%d", port)
			} else {
				startCmd = fmt.Sprintf("php -S 0.0.0.0:%d -t public", port)
			}
		}

	case hasFile("Gemfile"):
		if buildCmd == "" {
			buildCmd = "bundle install"
		}
		if startCmd == "" {
			startCmd = fmt.Sprintf("bundle exec ruby app.rb -p %d", port)
		}

	case hasFile("deno.json") || hasFile("deno.jsonc"):
		// Deno
		if startCmd == "" {
			entry := "main.ts"
			for _, e := range []string{"main.ts", "main.js", "src/main.ts", "src/index.ts"} {
				if hasFile(e) {
					entry = e
					break
				}
			}
			startCmd = fmt.Sprintf("deno run --allow-all %s --port %d", entry, port)
		}

	default:
		// Static HTML site or unknown project.
		if hasFile("index.html") {
			if startCmd == "" {
				if _, err := exec.LookPath("npx"); err == nil {
					startCmd = fmt.Sprintf("npx serve . -p %d", port)
				} else {
					startCmd = fmt.Sprintf("%s -m http.server %d", pythonExe, port)
				}
			}
		}
	}

	// Install dependencies
	if job.IsRecovery {
		w.appendLog(job.DeploymentID, "info", "system", "Recovery mode: skipping dependency installation")
	} else if job.InstallCommand != "" {
		w.appendLog(job.DeploymentID, "info", "system", fmt.Sprintf("Installing dependencies: %s", job.InstallCommand))
		shell, shellFlag := "sh", "-c"
		if runtime.GOOS == "windows" {
			shell, shellFlag = "cmd", "/c"
		}
		ic := exec.CommandContext(ctx, shell, shellFlag, job.InstallCommand)
		ic.Dir = workDir
		ic.Stdout = &logWriter{deploymentID: job.DeploymentID, stream: "stdout", w: w}
		ic.Stderr = &logWriter{deploymentID: job.DeploymentID, stream: "stderr", w: w}
		if err := ic.Run(); err != nil {
			return "", "", fmt.Errorf("install: %w", err)
		}
	} else if isNodeProject {
		w.appendLog(job.DeploymentID, "info", "system", fmt.Sprintf("Installing dependencies with %s...", pm))
		args := pmInstallArgs(pm)
		installCmd := exec.CommandContext(ctx, pm, args...)
		installCmd.Dir = workDir
		installCmd.Stdout = &logWriter{deploymentID: job.DeploymentID, stream: "stdout", w: w}
		installCmd.Stderr = &logWriter{deploymentID: job.DeploymentID, stream: "stderr", w: w}
		if err := installCmd.Run(); err != nil {
			return "", "", fmt.Errorf("%s install: %w", pm, err)
		}
	} else if isPythonProject {
		// Create a virtual environment so dependencies are isolated from the system.
		venvDir := filepath.Join(workDir, ".venv")
		w.appendLog(job.DeploymentID, "info", "system", "Creating Python virtual environment...")
		venvCmd := exec.CommandContext(ctx, pythonExe, "-m", "venv", ".venv")
		venvCmd.Dir = workDir
		venvCmd.Stdout = &logWriter{deploymentID: job.DeploymentID, stream: "stdout", w: w}
		venvCmd.Stderr = &logWriter{deploymentID: job.DeploymentID, stream: "stderr", w: w}
		if err := venvCmd.Run(); err != nil {
			w.appendLog(job.DeploymentID, "warn", "system",
				fmt.Sprintf("Could not create virtual environment (%v); installing into system Python", err))
			venvDir = ""
		}

		// Resolve pip and python paths; prefer venv when it was successfully created.
		pipBin := "pip"
		if venvDir != "" {
			if runtime.GOOS == "windows" {
				venvBinDir = filepath.Join(venvDir, "Scripts")
				pipBin = filepath.Join(venvBinDir, "pip.exe")
				pythonBin = filepath.Join(venvBinDir, "python.exe")
			} else {
				venvBinDir = filepath.Join(venvDir, "bin")
				pipBin = filepath.Join(venvBinDir, "pip")
				pythonBin = filepath.Join(venvBinDir, "python")
			}
		}

		// Upgrade pip first (best-effort).
		upCmd := exec.CommandContext(ctx, pythonBin, "-m", "pip", "install", "--upgrade", "pip", "-q")
		upCmd.Dir = workDir
		upCmd.Stdout = &logWriter{deploymentID: job.DeploymentID, stream: "stdout", w: w}
		upCmd.Stderr = &logWriter{deploymentID: job.DeploymentID, stream: "stderr", w: w}
		_ = upCmd.Run()

		w.appendLog(job.DeploymentID, "info", "system", "Installing Python dependencies with pip...")
		var pipArgs []string
		switch pythonReqFile {
		case "requirements.txt":
			pipArgs = []string{"install", "-r", "requirements.txt"}
		case "pyproject.toml", "setup.py":
			pipArgs = []string{"install", "-e", "."}
		case "Pipfile":
			// Use a requirements.txt alongside Pipfile if present; otherwise skip.
			if _, err := os.Stat(filepath.Join(workDir, "requirements.txt")); err == nil {
				pipArgs = []string{"install", "-r", "requirements.txt"}
			}
		}
		if len(pipArgs) > 0 {
			pipCmd := exec.CommandContext(ctx, pipBin, pipArgs...)
			pipCmd.Dir = workDir
			pipCmd.Stdout = &logWriter{deploymentID: job.DeploymentID, stream: "stdout", w: w}
			pipCmd.Stderr = &logWriter{deploymentID: job.DeploymentID, stream: "stderr", w: w}
			if err := pipCmd.Run(); err != nil {
				return "", "", fmt.Errorf("pip install: %w", err)
			}
		}

		// Point auto-generated startCmd at the venv python executable.
		if venvBinDir != "" && strings.HasPrefix(startCmd, pythonExe+" ") {
			startCmd = pythonBin + startCmd[len(pythonExe):]
		}
	}

	// Run build command
	if job.IsRecovery {
		w.appendLog(job.DeploymentID, "info", "system", "Recovery mode: skipping build step")
	} else if buildCmd != "" {
		w.appendLog(job.DeploymentID, "info", "system", fmt.Sprintf("Building: %s", buildCmd))
		shell, shellFlag := "sh", "-c"
		if runtime.GOOS == "windows" {
			shell, shellFlag = "cmd", "/c"
		}
		bc := exec.CommandContext(ctx, shell, shellFlag, buildCmd)
		bc.Dir = workDir
		bc.Stdout = &logWriter{deploymentID: job.DeploymentID, stream: "stdout", w: w}
		bc.Stderr = &logWriter{deploymentID: job.DeploymentID, stream: "stderr", w: w}
		if err := bc.Run(); err != nil {
			return "", "", fmt.Errorf("build: %w", err)
		}
	}

	// On Windows pnpm creates junction points inside node_modules/.pnpm that
	// cannot be read/copied with a naive filepath.Walk.  Instead of copying,
	// use the workDir as the run directory directly -- the process already has
	// all its dependencies there.  We symlink (or copy just the non-module
	// files) to a stable path under DeployDir so we can find the process later.
	runDir := workDir
	// Apply project-level run_dir subdirectory override if set
	if job.RunDir != "" {
		runDir = filepath.Join(workDir, job.RunDir)
		if _, err := os.Stat(runDir); err != nil {
			w.appendLog(job.DeploymentID, "warn", "system", fmt.Sprintf("run_dir '%s' not found, using repo root", job.RunDir))
			runDir = workDir
		}
	}
	// On Windows, the process is usually run from workDir (in CloneDir) because
	// of junction point issues. However, for session restore to work across
	// restarts where CloneDir might be wiped, we attempt to use DeployDir.
	permanentDir := filepath.Join(w.cfg.DeployDir, job.ProjectID[:8])

	if job.IsRecovery {
		w.appendLog(job.DeploymentID, "info", "system", "Recovery: attempting to use existing deployment directory")
		if _, err := os.Stat(permanentDir); err == nil {
			runDir = permanentDir
			if job.RunDir != "" {
				runDir = filepath.Join(permanentDir, job.RunDir)
			}
		}
	} else if runtime.GOOS != "windows" {
		// On Linux/macOS copy is fine (no Windows junctions).
		_ = os.RemoveAll(permanentDir)
		if err := copyDirSkipModules(workDir, permanentDir); err != nil {
			w.appendLog(job.DeploymentID, "warn", "system", fmt.Sprintf("Could not copy to run dir (%v) -- running from build dir", err))
		} else {
			runDir = permanentDir
			if job.RunDir != "" {
				runDir = filepath.Join(permanentDir, job.RunDir)
			}
		}
	} else {
		// On Windows (non-recovery): copy to permanentDir if possible so future restarts can recover
		_ = os.RemoveAll(permanentDir)
		if err := copyDirSkipModules(workDir, permanentDir); err == nil {
			runDir = permanentDir
			if job.RunDir != "" {
				runDir = filepath.Join(permanentDir, job.RunDir)
			}
		}
	}

	// Build a deduplicated environment for the process.
	// We start from os.Environ(), apply project env vars on top, then enforce
	// our own PORT and NODE_ENV last so they always win and never produce
	// the "non-standard NODE_ENV" warning from Next.js.
	envMap := make(map[string]string)
	for _, kv := range os.Environ() {
		if eq := strings.Index(kv, "="); eq > 0 {
			envMap[kv[:eq]] = kv[eq+1:]
		}
	}
	for k, v := range job.EnvVars {
		envMap[k] = v
	}
	// Enforce PORT; set NODE_ENV only for Node.js (other runtimes don't need it).
	envMap["PORT"] = fmt.Sprintf("%d", port)
	if isNodeProject {
		if _, userSet := job.EnvVars["NODE_ENV"]; !userSet {
			envMap["NODE_ENV"] = "production"
		}
	}
	// For Python venv: prepend bin dir to PATH so python/uvicorn/gunicorn resolve
	// to the venv-installed versions without needing the full path in startCmd.
	if venvBinDir != "" {
		pathSep := string(os.PathListSeparator)
		if existing, ok := envMap["PATH"]; ok {
			envMap["PATH"] = venvBinDir + pathSep + existing
		} else {
			envMap["PATH"] = venvBinDir
		}
		envMap["VIRTUAL_ENV"] = filepath.Dir(venvBinDir)
	}
	env := make([]string, 0, len(envMap))
	for k, v := range envMap {
		env = append(env, k+"="+v)
	}

	// If the user provided a start command with a hardcoded port that differs from
	// our assigned ExternalPort (common in zero-downtime), try to override it.
	if job.ExternalPort != 0 {
		// Replace common port flags: -p 3000, --port 3000, --port=3000
		re := regexp.MustCompile(`(?i)(--port[ =]|-p\s+)([0-9]{2,5})`)
		newStartCmd := re.ReplaceAllStringFunc(startCmd, func(match string) string {
			// Extract the prefix
			submatches := re.FindStringSubmatch(match)
			if len(submatches) >= 2 {
				return submatches[1] + fmt.Sprintf("%d", job.ExternalPort)
			}
			return match
		})
		if newStartCmd != startCmd {
			w.appendLog(job.DeploymentID, "info", "system", fmt.Sprintf("Adjusted start command port from %d to %d for zero-downtime", port, job.ExternalPort))
			startCmd = newStartCmd
		}
	}

	// Start the process
	w.appendLog(job.DeploymentID, "info", "system", fmt.Sprintf("Starting process: %s (ExternalPort: %d)", startCmd, port))
	shell, shellFlag := "sh", "-c"
	if runtime.GOOS == "windows" {
		shell, shellFlag = "cmd", "/c"
	}
	proc := exec.Command(shell, shellFlag, startCmd)
	proc.Dir = runDir
	proc.Env = env
	// Pipe output to logs asynchronously
	procLogger := &logWriter{deploymentID: job.DeploymentID, stream: "stdout", w: w}
	proc.Stdout = procLogger
	proc.Stderr = &logWriter{deploymentID: job.DeploymentID, stream: "stderr", w: w}

	if err := proc.Start(); err != nil {
		return "", "", fmt.Errorf("start process: %w", err)
	}

	processManager.Lock()
	processManager.procs[job.DeploymentID] = proc.Process
	processManager.Unlock()

	// Wait briefly to see if it crashes immediately.
	done := make(chan error, 1)
	go func() { done <- proc.Wait() }()
	select {
	case err := <-done:
		return "", "", fmt.Errorf("process exited immediately: %w", err)
	case <-time.After(3 * time.Second):
		// Still running -- good.
	}

	// Background goroutine: when the process eventually exits, update the
	// deployment status so the UI reflects the crash rather than staying "running".
	depID := job.DeploymentID
	go func() {
		err := <-done
		processManager.Lock()
		delete(processManager.procs, depID)
		processManager.Unlock()
		if err != nil {
			w.appendLog(depID, "error", "system", fmt.Sprintf("Process exited unexpectedly: %v", err))
			w.updateStatus(depID, "failed", err.Error())
		}
	}()

	deployURL := fmt.Sprintf("http://127.0.0.1:%d", port)

	// Health check and Zero-downtime swap
	if w.healthCheck(ctx, deployURL) {
		w.appendLog(job.DeploymentID, "info", "system", "Health check passed! Cleaning up old deployments...")
		// Mark current as running
		w.updateStatus(job.DeploymentID, "running", "")

		// Kill other 'running' deployments for this project
		var oldDeployments []models.Deployment
		w.db.Where("project_id = ? AND status = ? AND id != ?", job.ProjectID, "running", job.DeploymentID).Find(&oldDeployments)
		for _, old := range oldDeployments {
			w.appendLog(job.DeploymentID, "info", "system", fmt.Sprintf("Stopping old process: %s", old.ID))
			processManager.Lock()
			if p, ok := processManager.procs[old.ID]; ok {
				_ = p.Kill()
				delete(processManager.procs, old.ID)
			}
			processManager.Unlock()
			w.db.Model(&models.Deployment{}).Where("id = ?", old.ID).Update("status", "stopped")
		}
	} else {
		w.appendLog(job.DeploymentID, "error", "system", "Health check failed! Rolling back...")
		_ = proc.Process.Kill()
		processManager.Lock()
		delete(processManager.procs, job.DeploymentID)
		processManager.Unlock()
		return "", "", fmt.Errorf("health check failed")
	}

	return fmt.Sprintf("%d", proc.Process.Pid), deployURL, nil
}

// copyDirSkipModules recursively copies src to dst, skipping node_modules and .git.
func copyDirSkipModules(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// Skip entries we can't stat (e.g. Windows junction reparse points)
			return nil
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		// Skip entire node_modules and .git trees
		base := filepath.Base(rel)
		if info.IsDir() && (base == "node_modules" || base == ".git" || base == ".pnpm") {
			return filepath.SkipDir
		}
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, info.Mode())
		}
		// Skip symlinks on all platforms
		if info.Mode()&os.ModeSymlink != 0 {
			return nil
		}
		return copyFile(path, target, info.Mode())
	})
}

// copyDir recursively copies src to dst.
func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, info.Mode())
		}
		return copyFile(path, target, info.Mode())
	})
}

func copyFile(src, dst string, mode os.FileMode) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}
	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	_, err = io.Copy(out, in)
	// Ensure Close() error is not masked by Copy error
	if closeErr := out.Close(); closeErr != nil && err == nil {
		err = closeErr
	}
	return err
}

func (w *BuildWorker) cloneRepo(ctx context.Context, job *models.DeploymentJob, workDir string) error {
	cloneURL := job.RepoURL
	if job.GitToken != "" {
		// Embed PAT into the HTTPS URL for authenticated cloning:
		// https://github.com/user/repo  ->  https://<token>@github.com/user/repo
		if u, err := url.Parse(cloneURL); err == nil &&
			(u.Scheme == "https" || u.Scheme == "http") {
			u.User = url.User(job.GitToken)
			cloneURL = u.String()
		}
	}

	args := []string{"clone", "--depth=1", "--branch", job.Branch, cloneURL, workDir}
	if job.CommitSHA != "" {
		// For specific commits, we need full clone + checkout (no --depth)
		args = []string{"clone", job.RepoURL, workDir}
		if job.GitToken != "" {
			args = []string{"clone", cloneURL, workDir}
		}
	}

	// Run from the parent directory -- workDir must not exist yet for git clone.
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = filepath.Dir(workDir)
	out, err := cmd.CombinedOutput()

	if len(out) > 0 {
		// Redact the token from log output before appending
		logOut := out
		if job.GitToken != "" {
			logOut = []byte(strings.ReplaceAll(string(out), job.GitToken, "***"))
		}
		w.appendLog(job.DeploymentID, "info", "stdout", string(logOut))
	}

	if err != nil {
		return fmt.Errorf("git clone: %w", err)
	}

	// Checkout specific commit if provided
	if job.CommitSHA != "" {
		checkoutCmd := exec.CommandContext(ctx, "git", "checkout", job.CommitSHA)
		checkoutCmd.Dir = workDir
		if out, err := checkoutCmd.CombinedOutput(); err != nil {
			w.appendLog(job.DeploymentID, "warn", "stdout", string(out))
		} else {
			// Create a named rollback branch so the editor can reference it.
			branchName := "pushpaka/rollback-" + job.DeploymentID[:8]
			branchCmd := exec.CommandContext(ctx, "git", "checkout", "-b", branchName)
			branchCmd.Dir = workDir
			// Best-effort; ignore errors (branch may already exist).
			_, _ = branchCmd.CombinedOutput()
		}
	}
	return nil
}

func (w *BuildWorker) generateDockerfile(workDir string, job *models.DeploymentJob) error {
	var content string

	// Detect framework/language
	if _, err := os.Stat(filepath.Join(workDir, "package.json")); err == nil {
		pm := detectPackageManager(workDir)
		lockFile := pmLockFile(pm)
		installCI := pmCIArgs(pm)
		buildCmd := pm + " run build"
		startCmd := pm + " start"
		if job.BuildCommand != "" {
			buildCmd = job.BuildCommand
		}
		if job.StartCommand != "" {
			startCmd = job.StartCommand
		}
		// Install pm globally in the image if not npm (npm is pre-installed in node image)
		pmInstall := ""
		switch pm {
		case "pnpm":
			pmInstall = "RUN npm install -g pnpm\n"
		case "yarn":
			pmInstall = "RUN npm install -g yarn\n"
		case "bun":
			pmInstall = "RUN npm install -g bun\n"
		}
		content = fmt.Sprintf(`FROM node:20-alpine AS deps
WORKDIR /app
%sCOPY package.json %s ./
RUN %s

FROM node:20-alpine AS builder
WORKDIR /app
%sCOPY --from=deps /app/node_modules ./node_modules
COPY . .
RUN %s

FROM node:20-alpine AS runner
WORKDIR /app
ENV NODE_ENV production
COPY --from=builder /app .
EXPOSE %d
CMD [%s]
`, pmInstall, lockFile, installCI, pmInstall, buildCmd, job.Port, shellToCmdArray(startCmd))
	} else if _, err := os.Stat(filepath.Join(workDir, "go.mod")); err == nil {
		startCmd := "./app"
		if job.StartCommand != "" {
			startCmd = job.StartCommand
		}
		content = fmt.Sprintf(`FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o app .

FROM alpine:3.19
WORKDIR /app
COPY --from=builder /app/app .
EXPOSE %d
CMD ["%s"]
`, job.Port, startCmd)
	} else if _, err := os.Stat(filepath.Join(workDir, "requirements.txt")); err == nil {
		startCmd := "python app.py"
		if job.StartCommand != "" {
			startCmd = job.StartCommand
		}
		content = fmt.Sprintf(`FROM python:3.12-slim
WORKDIR /app
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt
COPY . .
EXPOSE %d
CMD [%s]
`, job.Port, shellToCmdArray(startCmd))
	} else if _, err := os.Stat(filepath.Join(workDir, "pyproject.toml")); err == nil {
		startCmd := "python -m uvicorn main:app --host 0.0.0.0 --port " + fmt.Sprintf("%d", job.Port)
		if job.StartCommand != "" {
			startCmd = job.StartCommand
		}
		content = fmt.Sprintf(`FROM python:3.12-slim
WORKDIR /app
RUN pip install --no-cache-dir uv
COPY pyproject.toml .
RUN uv pip install --system -e .
COPY . .
EXPOSE %d
CMD [%s]
`, job.Port, shellToCmdArray(startCmd))
	} else if _, err := os.Stat(filepath.Join(workDir, "Cargo.toml")); err == nil {
		// Rust project
		startCmd := "./app"
		if job.StartCommand != "" {
			startCmd = job.StartCommand
		} else if job.BuildCommand != "" {
			// infer binary name from build command
			startCmd = job.StartCommand
		}
		content = fmt.Sprintf(`FROM rust:1.76-slim AS builder
WORKDIR /app
COPY Cargo.toml Cargo.lock* ./
RUN mkdir src && echo 'fn main(){}' > src/main.rs && cargo build --release --locked 2>/dev/null; rm -rf src
COPY . .
RUN cargo build --release --locked

FROM debian:bookworm-slim
WORKDIR /app
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*
COPY --from=builder /app/target/release/* ./
EXPOSE %d
CMD ["%s"]
`, job.Port, startCmd)
	} else if _, err := os.Stat(filepath.Join(workDir, "index.html")); err == nil {
		// Static site — serve with nginx
		content = fmt.Sprintf(`FROM nginx:alpine
WORKDIR /usr/share/nginx/html
COPY . .
EXPOSE %d
CMD ["nginx", "-g", "daemon off;"]
`, job.Port)
	} else if _, err := os.Stat(filepath.Join(workDir, "Gemfile")); err == nil {
		// Ruby project
		startCmd := "bundle exec ruby app.rb"
		if job.StartCommand != "" {
			startCmd = job.StartCommand
		}
		content = fmt.Sprintf(`FROM ruby:3.3-slim
WORKDIR /app
COPY Gemfile Gemfile.lock* ./
RUN bundle install --without development test
COPY . .
EXPOSE %d
CMD [%s]
`, job.Port, shellToCmdArray(startCmd))
	} else {
		content = fmt.Sprintf(`FROM alpine:3.19
WORKDIR /app
COPY . .
EXPOSE %d
CMD ["./run.sh"]
`, job.Port)
	}

	return os.WriteFile(filepath.Join(workDir, "Dockerfile"), []byte(content), 0644)
}

func shellToCmdArray(cmd string) string {
	parts := strings.Fields(cmd)
	quoted := make([]string, len(parts))
	for i, p := range parts {
		quoted[i] = fmt.Sprintf("%q", p)
	}
	return strings.Join(quoted, ", ")
}

// detectPackageManager inspects lock files in workDir to choose the right
// package manager. Priority: bun > pnpm > yarn > npm (fallback).
// Returns the binary name ("npm", "yarn", "pnpm", "bun").
// NOTE: The caller is responsible for checking PATH availability; use
// exec.LookPath to fall back to npm when the returned binary is not installed.
func detectPackageManager(workDir string) string {
	if _, err := os.Stat(filepath.Join(workDir, "bun.lockb")); err == nil {
		return "bun"
	}
	if _, err := os.Stat(filepath.Join(workDir, "pnpm-lock.yaml")); err == nil {
		return "pnpm"
	}
	if _, err := os.Stat(filepath.Join(workDir, "yarn.lock")); err == nil {
		return "yarn"
	}
	return "npm"
}

// pmInstallArgs returns the install command arguments for the given package manager.
// Installs ALL dependencies (including devDeps) because the build step
// (e.g. `next build`, `vite build`) typically requires devDependencies.
func pmInstallArgs(pm string) []string {
	switch pm {
	case "pnpm":
		return []string{"install", "--frozen-lockfile"}
	case "yarn":
		return []string{"install", "--frozen-lockfile"}
	case "bun":
		return []string{"install"}
	default: // npm
		return []string{"install"}
	}
}

// pmCIArgs returns production-install args suitable for a Dockerfile (uses lockfile).
func pmCIArgs(pm string) string {
	switch pm {
	case "pnpm":
		return "pnpm install --prod --frozen-lockfile"
	case "yarn":
		return "yarn install --production --frozen-lockfile"
	case "bun":
		return "bun install --production"
	default: // npm
		return "npm ci --omit=dev"
	}
}

// pmLockFile returns the lock-file name for the given package manager.
func pmLockFile(pm string) string {
	switch pm {
	case "pnpm":
		return "pnpm-lock.yaml"
	case "yarn":
		return "yarn.lock"
	case "bun":
		return "bun.lockb"
	default:
		return "package-lock.json"
	}
}

// findPythonExe returns the Python executable name that is available in PATH.
// Prefers "python3" on Unix (common on modern systems), "python" on Windows.
func findPythonExe() string {
	if runtime.GOOS == "windows" {
		if _, err := exec.LookPath("python"); err == nil {
			return "python"
		}
		return "python3"
	}
	if _, err := exec.LookPath("python3"); err == nil {
		return "python3"
	}
	return "python"
}

// detectNodeStartCmd reads package.json and returns the best start command for
// the project.  It handles three common scenarios:
//  1. Vite / React + Vite: no  "start" script exists -- serve the built "dist" folder.
//  2. Express / generic Node: no "start" script -- fall back to "node <main>".
//  3. Everything else: use "<pm> start" (the standard behaviour).
func detectNodeStartCmd(workDir, pm string, port int) string {
	data, err := os.ReadFile(filepath.Join(workDir, "package.json"))
	if err != nil {
		return pm + " start"
	}
	var pkg struct {
		Main         string            `json:"main"`
		Scripts      map[string]string `json:"scripts"`
		Dependencies map[string]string `json:"dependencies"`
		DevDeps      map[string]string `json:"devDependencies"`
	}
	if err := json.Unmarshal(data, &pkg); err != nil {
		return pm + " start"
	}

	// If package.json explicitly declares a "start" script, use it.
	if _, hasStart := pkg.Scripts["start"]; hasStart {
		return pm + " start"
	}

	// No "start" script -- try to infer the correct command.
	allDeps := make(map[string]string)
	for k, v := range pkg.Dependencies {
		allDeps[k] = v
	}
	for k, v := range pkg.DevDeps {
		allDeps[k] = v
	}

	// Vite-based projects (React, Vue, Svelte, etc.) produce a "dist/" folder.
	if _, isVite := allDeps["vite"]; isVite {
		if _, err := exec.LookPath("npx"); err == nil {
			return fmt.Sprintf("npx serve dist -p %d", port)
		}
		return fmt.Sprintf("npx vite preview --port %d --host 0.0.0.0", port)
	}

	// Express / generic Node: use "node <main>" (default main: index.js).
	main := pkg.Main
	if main == "" {
		// Try common entry-point names.
		for _, name := range []string{"index.js", "server.js", "app.js", "src/index.js"} {
			if _, err := os.Stat(filepath.Join(workDir, name)); err == nil {
				main = name
				break
			}
		}
	}
	if main == "" {
		main = "index.js"
	}
	return "node " + main
}

// detectUvicornModule returns the "module:variable" string for uvicorn.
func detectUvicornModule(workDir string) string {
	for _, name := range []string{"main", "app", "server", "run", "asgi", "api"} {
		if _, err := os.Stat(filepath.Join(workDir, name+".py")); err == nil {
			return name + ":app"
		}
	}
	for _, sub := range []string{"src", "app", "backend", "api"} {
		for _, name := range []string{"main", "app", "server", "asgi"} {
			if _, err := os.Stat(filepath.Join(workDir, sub, name+".py")); err == nil {
				return sub + "." + name + ":app"
			}
		}
	}
	return "main:app"
}

// getRepoCommitInfo returns the HEAD commit SHA and subject line from git.
func getRepoCommitInfo(repoDir string) (sha, msg string, err error) {
	cmd := exec.Command("git", "log", "-1", "--format=%H|%s")
	cmd.Dir = repoDir
	out, err := cmd.Output()
	if err != nil {
		return "", "", err
	}
	parts := strings.SplitN(strings.TrimSpace(string(out)), "|", 2)
	if len(parts) == 2 {
		return parts[0], parts[1], nil
	}
	return strings.TrimSpace(string(out)), "", nil
}

// detectPythonEntry returns the best Python entry-point argument to pass to the
// interpreter (e.g. "app.py" or "manage.py runserver 0.0.0.0:3000").
// Falls back to "app.py" if nothing recognisable is found.
func detectPythonEntry(workDir string, port int) string {
	// Django: manage.py needs special arguments.
	if _, err := os.Stat(filepath.Join(workDir, "manage.py")); err == nil {
		return fmt.Sprintf("manage.py runserver 0.0.0.0:%d", port)
	}
	// FastAPI/Flask/generic script — look for common entry-point names.
	for _, name := range []string{"main.py", "app.py", "server.py", "run.py", "wsgi.py", "asgi.py", "index.py"} {
		if _, err := os.Stat(filepath.Join(workDir, name)); err == nil {
			return name
		}
	}
	// Check one level deep inside common sub-directories.
	for _, sub := range []string{"src", "app", "backend", "api"} {
		for _, name := range []string{"main.py", "app.py", "server.py"} {
			if _, err := os.Stat(filepath.Join(workDir, sub, name)); err == nil {
				return filepath.Join(sub, name)
			}
		}
	}
	return "app.py" // fallback
}

func (w *BuildWorker) buildImage(ctx context.Context, job *models.DeploymentJob, workDir string) error {
	// Use a per-project named cache volume so repeated builds reuse node_modules,
	// Go module cache, pip cache etc.  The cache volume is managed by Docker and
	// persists between builds.
	cacheMount := fmt.Sprintf("type=volume,source=pushpaka-cache-%s,target=/root/.cache", job.ProjectID[:8])
	args := []string{
		"build",
		"--cache-from", fmt.Sprintf("type=local,src=%s", buildCacheDir(w.cfg.CloneDir, job.ProjectID)),
		"--cache-to", fmt.Sprintf("type=local,dest=%s,mode=max", buildCacheDir(w.cfg.CloneDir, job.ProjectID)),
		"--mount", cacheMount,
		"-t", job.ImageTag,
		".",
	}
	// Fallback: if BuildKit is not available, use plain docker build without cache flags
	cmd := exec.CommandContext(ctx, "docker", args...)
	cmd.Dir = workDir
	cmd.Env = append(os.Environ(), "DOCKER_BUILDKIT=1")
	cmd.Stdout = &logWriter{deploymentID: job.DeploymentID, stream: "stdout", w: w}
	cmd.Stderr = &logWriter{deploymentID: job.DeploymentID, stream: "stderr", w: w}

	if err := cmd.Run(); err != nil {
		// Retry without BuildKit cache flags (older Docker versions)
		w.appendLog(job.DeploymentID, "warn", "system",
			"BuildKit cache failed, retrying without cache flags")
		plainArgs := []string{"build", "-t", job.ImageTag, "."}
		plain := exec.CommandContext(ctx, "docker", plainArgs...)
		plain.Dir = workDir
		plain.Stdout = &logWriter{deploymentID: job.DeploymentID, stream: "stdout", w: w}
		plain.Stderr = &logWriter{deploymentID: job.DeploymentID, stream: "stderr", w: w}
		return plain.Run()
	}
	return nil
}

// buildCacheDir returns the local directory used as a Docker build cache for a project.
func buildCacheDir(cloneDir, projectID string) string {
	return filepath.Join(cloneDir, ".buildcache", projectID[:8])
}

func (w *BuildWorker) deployContainer(ctx context.Context, job *models.DeploymentJob) (string, string, error) {
	// For zero-downtime, we don't kill the old container yet.
	// We use a unique name for the new container.
	containerName := fmt.Sprintf("pushpaka-%s-%s", job.ProjectID[:8], job.DeploymentID[:8])
	w.appendLog(job.DeploymentID, "info", "system", fmt.Sprintf("Starting new container: %s", containerName))

	// Build docker run arguments
	args := []string{
		"run", "-d",
		"--name", containerName,
		"--restart", "always",
		"--network", w.cfg.TraefikNetwork,
		"-p", fmt.Sprintf("%d:%d", job.ExternalPort, job.Port),
		// Traefik labels
		"--label", "traefik.enable=true",
		"--label", fmt.Sprintf("traefik.http.routers.%s.rule=PathPrefix(`/p/%s`)", containerName, job.ProjectID[:8]),
		"--label", fmt.Sprintf("traefik.http.services.%s.loadbalancer.server.port=%d", containerName, job.Port),
	}

	// Resource limits
	if job.CPULimit != "" {
		args = append(args, "--cpus="+job.CPULimit)
	}
	if job.MemoryLimit != "" {
		args = append(args, "--memory="+job.MemoryLimit)
		args = append(args, "--memory-swap="+job.MemoryLimit) // disable swap
	}

	// Add environment variables
	for k, v := range job.EnvVars {
		args = append(args, "-e", fmt.Sprintf("%s=%s", k, v))
	}

	args = append(args, job.ImageTag)

	cmd := exec.CommandContext(ctx, "docker", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", "", fmt.Errorf("docker run: %w\noutput: %s", err, string(out))
	}

	containerID := strings.TrimSpace(string(out))
	deployURL := fmt.Sprintf("http://localhost:%d", job.ExternalPort)

	// Health check and Zero-downtime swap
	if w.healthCheck(ctx, deployURL) {
		w.appendLog(job.DeploymentID, "info", "system", "Health check passed! Cleaning up old deployments...")
		// Mark current as running (it might have been building/queued)
		w.updateStatus(job.DeploymentID, "running", "")

		// Kill other 'running' deployments for this project
		var oldDeployments []models.Deployment
		w.db.Where("project_id = ? AND status = ? AND id != ?", job.ProjectID, "running", job.DeploymentID).Find(&oldDeployments)
		for _, old := range oldDeployments {
			w.appendLog(job.DeploymentID, "info", "system", fmt.Sprintf("Stopping old deployment: %s", old.ID))
			
			// Try the new naming scheme
			newOldContainerName := fmt.Sprintf("pushpaka-%s-%s", old.ProjectID[:8], old.ID[:8])
			_ = exec.CommandContext(ctx, "docker", "stop", newOldContainerName).Run()
			_ = exec.CommandContext(ctx, "docker", "rm", newOldContainerName).Run()

			// Fallback to legacy naming if needed
			legacyName := old.ProjectID[:8]
			_ = exec.CommandContext(ctx, "docker", "stop", legacyName).Run()
			_ = exec.CommandContext(ctx, "docker", "rm", legacyName).Run()

			w.db.Model(&models.Deployment{}).Where("id = ?", old.ID).Update("status", "stopped")
		}
	} else {
		w.appendLog(job.DeploymentID, "error", "system", "Health check failed! Rolling back...")
		_ = exec.CommandContext(ctx, "docker", "stop", containerName).Run()
		_ = exec.CommandContext(ctx, "docker", "rm", containerName).Run()
		return "", "", fmt.Errorf("health check failed")
	}

	return containerID, deployURL, nil
}

func (w *BuildWorker) fail(deploymentID, errMsg string) {
	log.Error().Str("deployment_id", deploymentID).Str("error", errMsg).Msg("deployment failed")
	w.appendLog(deploymentID, "error", "system", "FAILED: "+errMsg)
	w.updateStatus(deploymentID, "failed", errMsg)
}

// fireNotification calls the internal notification callback on the API server
// so that Slack/Discord/email alerts are fired without the worker needing
// direct access to those credentials.
func (w *BuildWorker) fireNotification(job *models.DeploymentJob, status, deployURL, errMsg string) {
	if job.NotificationURL == "" {
		return
	}

	// Fetch project name from DB for a better notification message.
	var projectName string
	w.db.Model(&models.Project{}).Where("id = ?", job.ProjectID).Pluck("name", &projectName)

	payload := map[string]any{
		"deployment_id": job.DeploymentID,
		"project_name":  projectName,
		"status":        status,
		"branch":        job.Branch,
		"commit_sha":    job.CommitSHA,
		"url":           deployURL,
		"error_msg":     errMsg,
		"user_id":       job.UserID,
	}
	data, _ := json.Marshal(payload)

	go func() {
		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Post(job.NotificationURL, "application/json", bytes.NewReader(data))
		if err != nil {
			log.Warn().Err(err).Str("url", job.NotificationURL).Msg("notification callback failed")
			return
		}
		resp.Body.Close()
	}()
}

func (w *BuildWorker) updateStatus(deploymentID, status, errMsg string) {
	err := w.db.Model(&models.Deployment{}).
		Where("id = ?", deploymentID).
		Updates(map[string]interface{}{
			"status":     status,
			"error_msg":  errMsg,
			"updated_at": time.Now().UTC(),
		}).Error
	if err != nil {
		log.Error().Err(err).Msg("failed to update deployment status")
	}
}

func (w *BuildWorker) appendLog(deploymentID, level, stream, message string) {
	err := w.db.Create(&models.DeploymentLog{
		BaseModel:    basemodel.BaseModel{ID: uuid.New().String()},
		DeploymentID: deploymentID,
		Level:        level,
		Stream:       stream,
		Message:      message,
	}).Error
	if err != nil {
		log.Error().Err(err).Msg("failed to append log")
	}
}

// logWriter streams docker build output to the DB
type logWriter struct {
	deploymentID string
	stream       string
	w            *BuildWorker
}

func (lw *logWriter) Write(p []byte) (int, error) {
	lines := strings.Split(strings.TrimSpace(string(p)), "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			lw.w.appendLog(lw.deploymentID, "info", lw.stream, line)
		}
	}
	return len(p), nil
}

func (w *BuildWorker) healthCheck(ctx context.Context, deployURL string) bool {
	w.appendLog("", "info", "system", fmt.Sprintf("Health check: %s", deployURL))
	// Give the app a moment to start
	time.Sleep(2 * time.Second)

	// Try for up to 30 seconds
	for i := 0; i < 15; i++ {
		select {
		case <-ctx.Done():
			return false
		default:
			// Just a simple GET request
			resp, err := http.Get(deployURL)
			if err == nil {
				resp.Body.Close()
				if resp.StatusCode < 500 {
					return true
				}
			}
			time.Sleep(2 * time.Second)
		}
	}
	return false
}
