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
	JobStarted(role string)
	JobFinished(role string)
}

// processManager tracks running direct-deployment processes (no Docker).
var processManager = struct {
	sync.Mutex
	procs map[string]*os.Process // deploymentID -> running process
}{procs: make(map[string]*os.Process)}

type BuildWorker struct {
	id              int
	db              *gorm.DB
	rdb             *redis.Client
	cfg             *config.Config
	dockerAvailable bool
	reporter        JobReporter
	Role            string
	Queue           string
}

func NewBuildWorker(id int, db *gorm.DB, rdb *redis.Client, cfg *config.Config, role, queue string) *BuildWorker {
	return &BuildWorker{
		id:              id,
		db:              db,
		rdb:             rdb,
		cfg:             cfg,
		dockerAvailable: checkDockerAvailable(cfg.DockerHost),
		Role:            role,
		Queue:           queue,
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
	log.Info().Int("worker_id", w.id).Str("role", w.Role).Msg("worker started")
	for {
		select {
		case <-ctx.Done():
			log.Info().Int("worker_id", w.id).Str("role", w.Role).Msg("worker stopping")
			return
		default:
			if w.rdb == nil {
				time.Sleep(1 * time.Second)
				continue
			}
			// Blocking pop from Redis queue with 5s timeout
			result, err := w.rdb.BRPop(ctx, 5*time.Second, w.Queue).Result()
			if err != nil {
				if err == redis.Nil {
					continue // timeout, try again
				}
				if ctx.Err() != nil {
					return // context cancelled
				}
				log.Error().Err(err).Str("role", w.Role).Msg("redis brpop error")
				continue
			}

			if len(result) < 2 {
				continue
			}

			taskID := result[1]
			w.processTask(ctx, taskID)
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
			taskID := string(payload)
			if reporter != nil {
				reporter.JobStarted(w.Role)
			}
			w.processTask(ctx, taskID)
			if reporter != nil {
				reporter.JobFinished(w.Role)
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

	// Update status based on job type
	initialStatus := string(models.DeploymentBuilding)
	if job.IsRecovery || job.IsBuildOnly {
		initialStatus = string(models.DeploymentRunning) // Or similar
		if job.IsBuildOnly {
			initialStatus = string(models.DeploymentBuilding)
		}
	}
	w.updateStatus(job.DeploymentID, initialStatus, "", "")
	w.appendLog(job.DeploymentID, "info", "system", "Worker process started")

	if !w.dockerAvailable {
		w.appendLog(job.DeploymentID, "info", "system", "Docker not available -- deploying directly (no containerization)")
	}

	// Source directory: versioned workspace if CommitSHA is present
	sourceDir := filepath.Join(w.cfg.ProjectsDir, job.ProjectID)
	if job.CommitSHA != "" {
		sourceDir = filepath.Join(w.cfg.ProjectsDir, job.ProjectID, job.CommitSHA)
	}
	
	// Build output directory: versioned storage for built artifacts.
	buildsDir := filepath.Join(w.cfg.BuildsDir, job.ProjectID)
	if job.CommitSHA != "" {
		buildsDir = filepath.Join(w.cfg.BuildsDir, job.ProjectID, job.CommitSHA)
	}

	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		w.fail(job.DeploymentID, fmt.Sprintf("failed to create source dir: %v", err))
		return
	}
	if err := os.MkdirAll(buildsDir, 0755); err != nil {
		w.fail(job.DeploymentID, fmt.Sprintf("failed to create builds dir: %v", err))
		return
	}

	// Step 0: Check for Build Cache
	if entries, _ := os.ReadDir(buildsDir); len(entries) > 0 {
		w.appendLog(job.DeploymentID, "info", "system", "Build artifacts found in cache (skipping build steps)")
		w.updateStatus(job.DeploymentID, string(models.DeploymentRunning), "", "") // Simplified: if build exists, assume we can skip to run or just finish build.
		// Update commit status to Built if not already
		w.updateCommitStatus(job.ProjectID, job.CommitSHA, models.CommitStatusBuilt)
		return
	}
	// Step 1: Clone or Sync repository
	needsClone := !job.IsRecovery
	if job.IsRecovery {
		// Verify if we can actually recover.
		canRecover := false
		if w.dockerAvailable {
			checkCmd := exec.CommandContext(ctx, "docker", "inspect", job.ImageTag)
			if err := checkCmd.Run(); err == nil {
				canRecover = true
			}
		} else {
			// For Direct: check if current deployment runtime dir exists.
			deployRuntimeDir := filepath.Join(w.cfg.DeploysDir, job.DeploymentID)
			if _, err := os.Stat(deployRuntimeDir); err == nil {
				canRecover = true
			}
		}

		if !canRecover {
			w.appendLog(job.DeploymentID, "warn", "system", "Recovery assets not found (image or directory) -- falling back to full build")
			needsClone = true
			job.IsRecovery = false
		} else {
			w.appendLog(job.DeploymentID, "info", "system", "Recovery mode: skipping repository sync")
		}
	}

	if needsClone {
		if _, err := os.Stat(filepath.Join(sourceDir, ".git")); os.IsNotExist(err) {
			w.appendLog(job.DeploymentID, "info", "system", fmt.Sprintf("Performing fresh clone to: %s", sourceDir))
			_ = os.RemoveAll(sourceDir)
			if err := w.cloneRepo(ctx, job, sourceDir); err != nil {
				w.fail(job.DeploymentID, fmt.Sprintf("clone failed: %v", err))
				return
			}
		} else {
			w.appendLog(job.DeploymentID, "info", "system", "Updating existing source code...")
			if err := w.syncRepo(ctx, job, sourceDir); err != nil {
				w.appendLog(job.DeploymentID, "warn", "system", fmt.Sprintf("Sync failed: %v. Attempting re-clone.", err))
				_ = os.RemoveAll(sourceDir)
				if err := w.cloneRepo(ctx, job, sourceDir); err != nil {
					w.fail(job.DeploymentID, fmt.Sprintf("re-clone failed: %v", err))
					return
				}
			}
		}
		w.appendLog(job.DeploymentID, "info", "system", "Repository synchronized successfully")
	}

	// Capture and update commit info (visible on project cards)
	if sha, msg, author, dateStr, err := getRepoCommitInfo(sourceDir); err == nil && sha != "" {
		job.CommitSHA = sha

		// Parse date
		var commitDate *time.Time
		if t, err := time.Parse("2006-01-02 15:04:05 -0700", dateStr); err == nil {
			commitDate = &t
		}

		// Update Project model for card visibility
		w.db.Model(&models.Project{}).Where("id = ?", job.ProjectID).Updates(map[string]interface{}{
			"latest_commit_sha": sha,
			"latest_commit_msg": msg,
			"latest_commit_at":  commitDate,
			"updated_at":         time.Now().UTC(),
		})
		
		w.appendLog(job.DeploymentID, "info", "system", fmt.Sprintf("Commit: %s by %s", sha[:7], author))
		// Update Deployment record
		w.db.Model(&models.Deployment{}).Where("id = ?", job.DeploymentID).Updates(map[string]interface{}{
			"commit_sha": sha,
			"commit_msg": msg,
		})
	}


	var containerID, deployURL string
	var deployErr error

	if w.dockerAvailable {
		// Docker path: generate Dockerfile -> build image -> run container
		dockerfilePath := filepath.Join(sourceDir, "Dockerfile")
		if _, err := os.Stat(dockerfilePath); os.IsNotExist(err) {
			w.appendLog(job.DeploymentID, "info", "system", "No Dockerfile found, generating one...")
			if err := w.generateDockerfile(sourceDir, job); err != nil {
				w.fail(job.DeploymentID, fmt.Sprintf("dockerfile generation failed: %v", err))
				return
			}
		}

		if job.IsRecovery {
			w.appendLog(job.DeploymentID, "info", "system", "Recovery mode: skipping image build")
		} else {
			w.appendLog(job.DeploymentID, "info", "system", fmt.Sprintf("Building Docker image: %s", job.ImageTag))
			if err := w.buildImage(ctx, job, sourceDir); err != nil {
				w.fail(job.DeploymentID, fmt.Sprintf("build failed: %v", err))
				return
			}
			w.appendLog(job.DeploymentID, "info", "system", "Docker image built successfully")
			// Update ProjectCommit status
			w.updateCommitStatus(job.ProjectID, job.CommitSHA, models.CommitStatusBuilt)
		}

		w.appendLog(job.DeploymentID, "info", "system", "Deploying container...")
		containerID, deployURL, deployErr = w.deployContainer(ctx, job)
	} else {
		// Direct path: install deps, build, then promote to BuildsDir, then deploy
		w.appendLog(job.DeploymentID, "info", "system", "Preparing project for direct deployment...")
		
		// 1. Build in sourceDir
		w.appendLog(job.DeploymentID, "info", "system", "Installing dependencies and building in source directory...")
		if err := w.runBuildInSource(ctx, job, sourceDir); err != nil {
			w.fail(job.DeploymentID, fmt.Sprintf("build failed: %v", err))
			return
		}
		// Update ProjectCommit status
		w.updateCommitStatus(job.ProjectID, job.CommitSHA, models.CommitStatusBuilt)

		// 2. Promote to buildsDir
		w.appendLog(job.DeploymentID, "info", "system", "Promoting build artifacts to builds directory...")
		if err := w.promoteToBuildsDir(sourceDir, buildsDir); err != nil {
			w.appendLog(job.DeploymentID, "warn", "system", fmt.Sprintf("Promotion failed: %v. Using source dir directly.", err))
			// Non-fatal, but we'll try to deploy from source if promotion fails
		}

		// 3. Deploy from buildsDir (or sourceDir if promotion failed)
		if job.IsBuildOnly {
			w.appendLog(job.DeploymentID, "success", "system", "Build completed successfully (Build-only mode).")
			w.updateStatus(job.DeploymentID, "finished", "", "") // Or some other final state
			return
		}

		deployBaseDir := buildsDir
		if _, err := os.Stat(buildsDir); os.IsNotExist(err) {
			deployBaseDir = sourceDir
		}

		containerID, deployURL, deployErr = w.deployDirect(ctx, job, deployBaseDir)
	}

	if deployErr != nil {
		msg := fmt.Sprintf("deployment failed: %v", deployErr)
		w.fail(job.DeploymentID, msg)
		w.fireNotification(job, "failed", "", msg)
		return
	}

	// Persist source code to the permanent deployment directory so the editor can see it.
	// This is done for both Docker and Direct deployments.
	permanentDir := filepath.Join(w.cfg.DeploysDir, job.ProjectID[:8])
	if !job.IsRecovery {
		w.appendLog(job.DeploymentID, "info", "system", "Persisting source code for editor...")
		_ = os.MkdirAll(filepath.Dir(permanentDir), 0755)
		_ = os.RemoveAll(permanentDir)
		if err := copyDirSkipModules(sourceDir, permanentDir); err != nil {
			w.appendLog(job.DeploymentID, "warn", "system", fmt.Sprintf("failed to persist source for editor: %v", err))
		}
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
func (w *BuildWorker) deployDirect(ctx context.Context, job *models.DeploymentJob, deployBaseDir string) (string, string, error) {
	// For zero-downtime, we don't kill the old process yet.
	// The new process will start on a separate port (job.ExternalPort).
	sourceDir := filepath.Join(w.cfg.ProjectsDir, job.ProjectID)

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
	_ = pythonReqFile   // avoid unused lint
	pythonExe := findPythonExe()
	var venvBinDir, pythonBin string
	pythonBin = pythonExe

	hasFile := func(name string) bool {
		_, err := os.Stat(filepath.Join(sourceDir, name))
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
		pm = detectPackageManager(sourceDir)
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
		if pkgData, err := os.ReadFile(filepath.Join(sourceDir, "package.json")); err == nil {
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
					_ = os.WriteFile(filepath.Join(sourceDir, "next.config.ts"), []byte(tsContent), 0644)
					w.appendLog(job.DeploymentID, "info", "system", "Created minimal next.config.ts (TypeScript project, no config found)")
				} else {
					// Check for ESM package (\"type\": \"module\")
					isESM := false
					if pkgData, err := os.ReadFile(filepath.Join(sourceDir, "package.json")); err == nil {
						s := string(pkgData)
						isESM = strings.Contains(s, `"type": "module"`) || strings.Contains(s, `"type":"module"`)
					}
					if isESM {
						esmContent := "/** @type {import('next').NextConfig} */\nconst nextConfig = {}\nexport default nextConfig\n"
						_ = os.WriteFile(filepath.Join(sourceDir, "next.config.mjs"), []byte(esmContent), 0644)
						w.appendLog(job.DeploymentID, "info", "system", "Created minimal next.config.mjs (ESM project, no config found)")
					} else {
						jsContent := "/** @type {import('next').NextConfig} */\nconst nextConfig = {}\nmodule.exports = nextConfig\n"
						_ = os.WriteFile(filepath.Join(sourceDir, "next.config.js"), []byte(jsContent), 0644)
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
			startCmd = detectNodeStartCmd(sourceDir, pm, port)
		}

	case hasFile("requirements.txt") || isPrimaryPython:
		isPythonProject = true
		pythonReqFile = "requirements.txt"
		if startCmd == "" {
			// Check if FastAPI/uvicorn is listed as a dependency.
			if reqBytes, err := os.ReadFile(filepath.Join(sourceDir, "requirements.txt")); err == nil {
				reqLower := strings.ToLower(string(reqBytes))
				if strings.Contains(reqLower, "fastapi") || strings.Contains(reqLower, "uvicorn") ||
					strings.Contains(reqLower, "starlette") {
					startCmd = fmt.Sprintf("uvicorn %s --host 0.0.0.0 --port %d",
						detectUvicornModule(sourceDir), port)
				}
			}
			if startCmd == "" {
				startCmd = pythonExe + " " + detectPythonEntry(sourceDir, port)
			}
		}

	case hasFile("pyproject.toml"):
		isPythonProject = true
		pythonReqFile = "pyproject.toml"
		if startCmd == "" {
			// Check if this is a FastAPI/uvicorn project.
			if content, err := os.ReadFile(filepath.Join(sourceDir, "pyproject.toml")); err == nil {
				s := string(content)
				if strings.Contains(s, "fastapi") || strings.Contains(s, "uvicorn") {
					startCmd = fmt.Sprintf("uvicorn main:app --host 0.0.0.0 --port %d", port)
				}
			}
			if startCmd == "" {
				startCmd = pythonExe + " " + detectPythonEntry(sourceDir, port)
			}
		}

	case hasFile("Pipfile"):
		isPythonProject = true
		pythonReqFile = "Pipfile"
		if startCmd == "" {
			startCmd = pythonExe + " " + detectPythonEntry(sourceDir, port)
		}

	case hasFile("setup.py"):
		isPythonProject = true
		pythonReqFile = "setup.py"
		if startCmd == "" {
			startCmd = pythonExe + " " + detectPythonEntry(sourceDir, port)
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
				_ = os.Chmod(filepath.Join(sourceDir, "gradlew"), 0755)
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

	// ─── Runtime context ─────────────────────────────────────────────────────
	deploymentDir := filepath.Join(w.cfg.DeploysDir, job.DeploymentID)
	if err := os.MkdirAll(deploymentDir, 0755); err != nil {
		return "", "", fmt.Errorf("failed to create deployment dir: %v", err)
	}

	// For direct deployments, we're settled on running from the isolated deployment dir.
	if job.IsRecovery {
		w.appendLog(job.DeploymentID, "info", "system", "Recovery mode: skipping artifact copy")
	} else {
		w.appendLog(job.DeploymentID, "info", "system", "Copying artifacts to isolated deployment directory...")
		if err := copyDir(deployBaseDir, deploymentDir); err != nil {
			return "", "", fmt.Errorf("failed to copy artifacts to deployment dir: %v", err)
		}
	}

	runDir := deploymentDir
	if job.RunDir != "" {
		runDir = filepath.Join(deploymentDir, job.RunDir)
	}

	// Build environment map
	whitelist := []string{"PATH", "HOME", "USER", "LANG", "SystemRoot", "SystemDrive", "TEMP", "TMP"}
	envMap := make(map[string]string)
	for _, key := range whitelist {
		if val := os.Getenv(key); val != "" {
			envMap[key] = val
		}
	}
	for k, v := range job.EnvVars {
		envMap[k] = v
	}
	envMap["PORT"] = fmt.Sprintf("%d", port)
	if isNodeProject {
		if _, userSet := job.EnvVars["NODE_ENV"]; !userSet {
			envMap["NODE_ENV"] = "production"
		}
	}

	// Python venv support in isolated dir
	if isPythonProject {
		venvDir := filepath.Join(runDir, ".venv")
		// If venv exists in buildsDir, it was copied. If not, maybe we need it?
		if _, err := os.Stat(venvDir); err == nil {
			if runtime.GOOS == "windows" {
				venvBinDir = filepath.Join(venvDir, "Scripts")
				pythonBin = filepath.Join(venvBinDir, "python.exe")
				// Fallback if Scripts not found but bin exists
				if _, err := os.Stat(pythonBin); err != nil {
					if _, err := os.Stat(filepath.Join(venvDir, "bin", "python.exe")); err == nil {
						venvBinDir = filepath.Join(venvDir, "bin")
						pythonBin = filepath.Join(venvBinDir, "python.exe")
					}
				}
			} else {
				venvBinDir = filepath.Join(venvDir, "bin")
				pythonBin = filepath.Join(venvBinDir, "python")
			}
			
			if venvBinDir != "" {
				pathSep := string(os.PathListSeparator)
				if existing, ok := envMap["PATH"]; ok {
					envMap["PATH"] = venvBinDir + pathSep + existing
				} else {
					envMap["PATH"] = venvBinDir
				}
				envMap["VIRTUAL_ENV"] = venvDir
				
				// Update startCmd to use venv python
				pythonExe := findPythonExe()
				if strings.HasPrefix(startCmd, pythonExe+" ") {
					startCmd = pythonBin + startCmd[len(pythonExe):]
				}
			}
		}
	}

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

	// Write .env file
	if len(job.EnvVars) > 0 {
		var envFileContent strings.Builder
		for k, v := range job.EnvVars {
			envFileContent.WriteString(fmt.Sprintf("%s=%s\n", k, v))
		}
		_ = os.WriteFile(filepath.Join(runDir, ".env"), []byte(envFileContent.String()), 0600)
	}

	w.appendLog(job.DeploymentID, "info", "system", fmt.Sprintf("Starting process: %s (ExternalPort: %d)", startCmd, port))
	shell, shellFlag := "sh", "-c"
	if runtime.GOOS == "windows" {
		shell, shellFlag = "cmd", "/c"
	}
	proc := exec.Command(shell, shellFlag, startCmd)
	proc.Dir = runDir
	proc.Env = env
	proc.Stdout = &logWriter{deploymentID: job.DeploymentID, stream: "stdout", w: w}
	proc.Stderr = &logWriter{deploymentID: job.DeploymentID, stream: "stderr", w: w}

	if err := proc.Start(); err != nil {
		return "", "", fmt.Errorf("start process: %w", err)
	}

	processManager.Lock()
	processManager.procs[job.DeploymentID] = proc.Process
	processManager.Unlock()

	// Monitoring...
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

		// If the context is cancelled, the worker is shutting down.
		// We don't mark as "failed" because the exit was likely induced by the shutdown signal.
		// Keeping status as "running" allows the service to recover it on next startup.
		if ctx.Err() != nil {
			w.appendLog(depID, "info", "system", "Worker shutting down: process stopped (will recover on restart)")
			return
		}

		if err != nil {
			w.appendLog(depID, "error", "system", fmt.Sprintf("Process exited unexpectedly: %v", err))
			w.updateStatus(depID, string(models.DeploymentFailed), err.Error(), "")
		} else {
			// Normal exit (status 0) - mark as stopped so the UI is accurate.
			w.appendLog(depID, "info", "system", "Process exited normally")
			w.updateStatus(depID, string(models.DeploymentStopped), "", "")
		}
	}()

	deployURL := fmt.Sprintf("http://127.0.0.1:%d", port)

	// Health check and Zero-downtime swap
	if w.healthCheck(ctx, deployURL) {
		w.appendLog(job.DeploymentID, "info", "system", "Health check passed! Cleaning up old deployments...")
		// Mark current as running
		w.updateStatus(job.DeploymentID, string(models.DeploymentRunning), "", "")

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
			w.db.Model(&models.Deployment{}).Where("id = ?", old.ID).Update("status", string(models.DeploymentStopped))
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

func (w *BuildWorker) cloneRepo(ctx context.Context, job *models.DeploymentJob, sourceDir string) error {
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

	args := []string{"clone", "--depth=1", "--branch", job.Branch, cloneURL, sourceDir}
	if job.CommitSHA != "" {
		// For specific commits, we need full clone + checkout (no --depth)
		args = []string{"clone", job.RepoURL, sourceDir}
		if job.GitToken != "" {
			args = []string{"clone", cloneURL, sourceDir}
		}
	}

	// Run from the parent directory -- sourceDir must not exist yet for git clone.
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = filepath.Dir(sourceDir)
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
		checkoutCmd.Dir = sourceDir
		if out, err := checkoutCmd.CombinedOutput(); err != nil {
			w.appendLog(job.DeploymentID, "warn", "stdout", string(out))
		} else {
			// Create a named rollback branch so the editor can reference it.
			branchName := "pushpaka/rollback-" + job.DeploymentID[:8]
			branchCmd := exec.CommandContext(ctx, "git", "checkout", "-b", branchName)
			branchCmd.Dir = sourceDir
			// Best-effort; ignore errors (branch may already exist).
			_, _ = branchCmd.CombinedOutput()
		}
	}
	return nil
}

// syncRepo performs a fetch and hard reset to ensure the local source matches the remote.
func (w *BuildWorker) syncRepo(ctx context.Context, job *models.DeploymentJob, sourceDir string) error {
	w.appendLog(job.DeploymentID, "info", "system", "Synchronizing repository changes...")
	
	// 1. Fetch
	fetchCmd := exec.CommandContext(ctx, "git", "fetch", "--all")
	fetchCmd.Dir = sourceDir
	if out, err := fetchCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git fetch: %v: %s", err, string(out))
	}

	// 2. Reset hard to the target branch
	target := "origin/" + job.Branch
	resetCmd := exec.CommandContext(ctx, "git", "reset", "--hard", target)
	resetCmd.Dir = sourceDir
	if out, err := resetCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git reset: %v: %s", err, string(out))
	}

	// 3. Clean untracked files (optional but keeps it clean)
	cleanCmd := exec.CommandContext(ctx, "git", "clean", "-fd")
	cleanCmd.Dir = sourceDir
	_ = cleanCmd.Run()

	return nil
}

func (w *BuildWorker) generateDockerfile(sourceDir string, job *models.DeploymentJob) error {
	var content string

	// Detect framework/language
	if _, err := os.Stat(filepath.Join(sourceDir, "package.json")); err == nil {
		pm := detectPackageManager(sourceDir)
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
	} else if _, err := os.Stat(filepath.Join(sourceDir, "go.mod")); err == nil {
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
	} else if _, err := os.Stat(filepath.Join(sourceDir, "requirements.txt")); err == nil {
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
	} else if _, err := os.Stat(filepath.Join(sourceDir, "pyproject.toml")); err == nil {
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
	} else if _, err := os.Stat(filepath.Join(sourceDir, "Cargo.toml")); err == nil {
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
	} else if _, err := os.Stat(filepath.Join(sourceDir, "index.html")); err == nil {
		// Static site — serve with nginx
		content = fmt.Sprintf(`FROM nginx:alpine
WORKDIR /usr/share/nginx/html
COPY . .
EXPOSE %d
CMD ["nginx", "-g", "daemon off;"]
`, job.Port)
	} else if _, err := os.Stat(filepath.Join(sourceDir, "Gemfile")); err == nil {
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

	return os.WriteFile(filepath.Join(sourceDir, "Dockerfile"), []byte(content), 0644)
}

func shellToCmdArray(cmd string) string {
	parts := strings.Fields(cmd)
	quoted := make([]string, len(parts))
	for i, p := range parts {
		quoted[i] = fmt.Sprintf("%q", p)
	}
	return strings.Join(quoted, ", ")
}

// detectPackageManager inspects lock files in sourceDir to choose the right
// package manager. Priority: bun > pnpm > yarn > npm (fallback).
// Returns the binary name ("npm", "yarn", "pnpm", "bun").
// NOTE: The caller is responsible for checking PATH availability; use
// exec.LookPath to fall back to npm when the returned binary is not installed.
func detectPackageManager(sourceDir string) string {
	if _, err := os.Stat(filepath.Join(sourceDir, "bun.lockb")); err == nil {
		return "bun"
	}
	if _, err := os.Stat(filepath.Join(sourceDir, "pnpm-lock.yaml")); err == nil {
		return "pnpm"
	}
	if _, err := os.Stat(filepath.Join(sourceDir, "yarn.lock")); err == nil {
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
func detectNodeStartCmd(sourceDir, pm string, port int) string {
	data, err := os.ReadFile(filepath.Join(sourceDir, "package.json"))
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
			if _, err := os.Stat(filepath.Join(sourceDir, name)); err == nil {
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
func detectUvicornModule(sourceDir string) string {
	for _, name := range []string{"main", "app", "server", "run", "asgi", "api"} {
		if _, err := os.Stat(filepath.Join(sourceDir, name+".py")); err == nil {
			return name + ":app"
		}
	}
	for _, sub := range []string{"src", "app", "backend", "api"} {
		for _, name := range []string{"main", "app", "server", "asgi"} {
			if _, err := os.Stat(filepath.Join(sourceDir, sub, name+".py")); err == nil {
				return sub + "." + name + ":app"
			}
		}
	}
	return "main:app"
}

// getRepoCommitInfo returns the HEAD commit SHA and subject line from git.
func getRepoCommitInfo(repoDir string) (sha, msg, author, date string, err error) {
	// format: hash|subject|author_name|author_date(iso8601)
	cmd := exec.Command("git", "log", "-1", "--format=%H|%s|%an|%ai")
	cmd.Dir = repoDir
	out, err := cmd.Output()
	if err != nil {
		return "", "", "", "", err
	}
	parts := strings.SplitN(strings.TrimSpace(string(out)), "|", 4)
	if len(parts) == 4 {
		return parts[0], parts[1], parts[2], parts[3], nil
	}
	if len(parts) > 0 {
		return parts[0], "", "", "", nil
	}
	return "", "", "", "", fmt.Errorf("unexpected git log output")
}

// detectPythonEntry returns the best Python entry-point argument to pass to the
// interpreter (e.g. "app.py" or "manage.py runserver 0.0.0.0:3000").
// Falls back to "app.py" if nothing recognisable is found.
func detectPythonEntry(sourceDir string, port int) string {
	// Django: manage.py needs special arguments.
	if _, err := os.Stat(filepath.Join(sourceDir, "manage.py")); err == nil {
		return fmt.Sprintf("manage.py runserver 0.0.0.0:%d", port)
	}
	// FastAPI/Flask/generic script — look for common entry-point names.
	for _, name := range []string{"main.py", "app.py", "server.py", "run.py", "wsgi.py", "asgi.py", "index.py"} {
		if _, err := os.Stat(filepath.Join(sourceDir, name)); err == nil {
			return name
		}
	}
	// Check one level deep inside common sub-directories.
	for _, sub := range []string{"src", "app", "backend", "api"} {
		for _, name := range []string{"main.py", "app.py", "server.py"} {
			if _, err := os.Stat(filepath.Join(sourceDir, sub, name)); err == nil {
				return filepath.Join(sub, name)
			}
		}
	}
	return "app.py" // fallback
}

func (w *BuildWorker) buildImage(ctx context.Context, job *models.DeploymentJob, sourceDir string) error {
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
	cmd.Dir = sourceDir
	cmd.Env = append(os.Environ(), "DOCKER_BUILDKIT=1")
	cmd.Stdout = &logWriter{deploymentID: job.DeploymentID, stream: "stdout", w: w}
	cmd.Stderr = &logWriter{deploymentID: job.DeploymentID, stream: "stderr", w: w}

	if err := cmd.Run(); err != nil {
		// Retry without BuildKit cache flags (older Docker versions)
		w.appendLog(job.DeploymentID, "warn", "system",
			"BuildKit cache failed, retrying without cache flags")
		plainArgs := []string{"build", "-t", job.ImageTag, "."}
		plain := exec.CommandContext(ctx, "docker", plainArgs...)
		plain.Dir = sourceDir
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
		w.updateStatus(job.DeploymentID, string(models.DeploymentRunning), "", "")

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

func (w *BuildWorker) fail(id, errMsg string) {
	log.Error().Str("id", id).Str("error", errMsg).Msg("task or deployment failed")
	w.appendLog(id, "error", "system", "FAILED: "+errMsg)
	
	resolution := ""
	if w.cfg.AIAPIKey != "" {
		w.appendLog(id, "info", "system", "AI Assistant analyzing failure for immediate resolution...")
		resolution = w.analyzeFailure(id, errMsg)
		if resolution != "" {
			w.appendLog(id, "info", "system", "AI RECOMMENDED FIX: " + resolution)
		}
	}

	w.updateStatus(id, "failed", errMsg, resolution)
	// Explicitly notify completion for tasks if this was a task
	w.completeTask(id, false, errMsg)
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

func (w *BuildWorker) updateStatus(id, status, errMsg, resolution string) {
	now := time.Now().UTC()
	
	// 1. Try to update Deployment record
	res := w.db.Model(&models.Deployment{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":     status,
			"error_msg":  errMsg,
			"resolution": resolution,
			"updated_at": now,
		})
	
	if res.Error == nil && res.RowsAffected > 0 {
		// If we're starting (building or running) and haven't recorded StartedAt yet, do it now.
		if status == string(models.DeploymentBuilding) || status == string(models.DeploymentRunning) {
			w.db.Model(&models.Deployment{}).
				Where("id = ? AND started_at IS NULL", id).
				Update("started_at", now)
		}
		return
	}

	// 2. Try to update ProjectTask record if no deployment was updated
	taskStatus := models.TaskStatus(status)
	// Map deployment statuses to task statuses if needed
	if status == "failed" {
		taskStatus = models.TaskStatusFailed
	} else if status == "running" || status == "building" {
		taskStatus = models.TaskStatusRunning
	} else if status == "completed" || status == "finished" {
		taskStatus = models.TaskStatusCompleted
	}

	w.db.Model(&models.ProjectTask{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":      taskStatus,
			"error":       errMsg,
			"finished_at": now,
		})
}

func (w *BuildWorker) updateCommitStatus(projectID, sha string, status models.CommitStatus) {
	if sha == "" {
		return
	}
	err := w.db.Model(&models.ProjectCommit{}).
		Where("project_id = ? AND sha = ?", projectID, sha).
		Update("status", status).Error
	if err != nil {
		log.Error().Err(err).Str("project_id", projectID).Str("sha", sha).Msg("failed to update commit status")
	}
}

func (w *BuildWorker) analyzeFailure(deploymentID, errMsg string) string {
	// Fetch last 50 logs for context
	var logs []models.DeploymentLog
	w.db.Where("deployment_id = ?", deploymentID).Order("created_at desc").Limit(50).Find(&logs)
	
	var sb strings.Builder
	for i := len(logs)-1; i >= 0; i-- {
		sb.WriteString(logs[i].Message + "\n")
	}
	contextLogs := sb.String()

	prompt := fmt.Sprintf(`The deployment failed with error: %s
Recent logs:
%s
Please provide a very short, one-sentence resolution or fix for this issue.`, errMsg, contextLogs)

	payload := map[string]any{
		"model": w.cfg.AIModel,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
	}
	
	data, _ := json.Marshal(payload)
	
	apiURL := w.cfg.AIBaseURL
	if apiURL == "" {
		switch w.cfg.AIProvider {
		case "openai": apiURL = "https://api.openai.com/v1/chat/completions"
		case "openrouter": apiURL = "https://openrouter.ai/api/v1/chat/completions"
		case "ollama": apiURL = "http://localhost:11434/api/chat"
		default: return ""
		}
	}

	client := &http.Client{Timeout: 30 * time.Second}
	req, _ := http.NewRequest("POST", apiURL, bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	if w.cfg.AIAPIKey != "" {
		req.Header.Set("Authorization", "Bearer "+w.cfg.AIAPIKey)
	}
	
	resp, err := client.Do(req)
	if err != nil {
		log.Warn().Err(err).Msg("AI analysis failed to connect")
		return ""
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return ""
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err == nil && len(result.Choices) > 0 {
		return strings.TrimSpace(result.Choices[0].Message.Content)
	}
	
	return ""
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
				// Only consider 2xx (Success) or 3xx (Redirect) as healthy.
				// 4xx (Client Error) or 5xx (Server Error) are failures.
				if resp.StatusCode >= 200 && resp.StatusCode < 400 {
					return true
				}
				w.appendLog("", "warn", "system", fmt.Sprintf("Health check returned status: %d (not ready)", resp.StatusCode))
			}
			time.Sleep(2 * time.Second)
		}
	}
	return false
}

// runBuildInSource executes dependency installation and build scripts in the source directory.
func (w *BuildWorker) runBuildInSource(ctx context.Context, job *models.DeploymentJob, sourceDir string) error {
	if _, err := os.Stat(filepath.Join(sourceDir, "package.json")); err == nil {
		pm := detectPackageManager(sourceDir)
		
		// Install
		installCmd := pm + " install"
		if job.InstallCommand != "" {
			installCmd = job.InstallCommand
		}
		w.appendLog(job.DeploymentID, "info", "system", "Running install: "+installCmd)
		cmd := exec.CommandContext(ctx, "sh", "-c", installCmd)
		if runtime.GOOS == "windows" {
			cmd = exec.CommandContext(ctx, "cmd", "/c", installCmd)
		}
		cmd.Dir = sourceDir
		// Add node_modules/.bin to PATH for Windows/Direct builds
		pathSep := string(os.PathListSeparator)
		binPath := filepath.Join(sourceDir, "node_modules", ".bin")
		cmd.Env = append(os.Environ(), "PATH="+binPath+pathSep+os.Getenv("PATH"))
		
		cmd.Stdout = &logWriter{deploymentID: job.DeploymentID, stream: "stdout", w: w}
		cmd.Stderr = &logWriter{deploymentID: job.DeploymentID, stream: "stderr", w: w}
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("install failed: %v", err)
		}

		// Build
		if job.BuildCommand != "" || hasBuildScript(sourceDir) {
			buildCmd := pm + " run build"
			if job.BuildCommand != "" {
				buildCmd = job.BuildCommand
			}
			w.appendLog(job.DeploymentID, "info", "system", "Running build: "+buildCmd)
			cmd = exec.CommandContext(ctx, "sh", "-c", buildCmd)
			if runtime.GOOS == "windows" {
				cmd = exec.CommandContext(ctx, "cmd", "/c", buildCmd)
			}
			cmd.Dir = sourceDir
			binPath := filepath.Join(sourceDir, "node_modules", ".bin")
			pathSep := string(os.PathListSeparator)
			cmd.Env = append(os.Environ(), "PATH="+binPath+pathSep+os.Getenv("PATH"))
			
			cmd.Stdout = &logWriter{deploymentID: job.DeploymentID, stream: "stdout", w: w}
			cmd.Stderr = &logWriter{deploymentID: job.DeploymentID, stream: "stderr", w: w}
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("build failed: %v", err)
			}
		}
	} else if _, err := os.Stat(filepath.Join(sourceDir, "requirements.txt")); err == nil {
		// Python build (optional)
		w.appendLog(job.DeploymentID, "info", "system", "Python project detected, skipping build step.")
	}
	return nil
}

// promoteToBuildsDir copies build artifacts from sourceDir to buildsDir.
func (w *BuildWorker) promoteToBuildsDir(sourceDir, buildsDir string) error {
	_ = os.RemoveAll(buildsDir)
	_ = os.MkdirAll(buildsDir, 0755)

	// Artifact selection: common output directories.
	artifactDir := ""
	for _, dir := range []string{"dist", "build", ".next", "out", "public"} {
		if _, err := os.Stat(filepath.Join(sourceDir, dir)); err == nil {
			artifactDir = dir
			break
		}
	}

	if artifactDir != "" {
		src := filepath.Join(sourceDir, artifactDir)
		return copyDir(src, buildsDir)
	}

	// Fallback: copy whole source but skip node_modules etc.
	return copyDirSkipModules(sourceDir, buildsDir)
}

func hasBuildScript(dir string) bool {
	data, err := os.ReadFile(filepath.Join(dir, "package.json"))
	if err != nil {
		return false
	}
	var pkg struct {
		Scripts map[string]string `json:"scripts"`
	}
	if err := json.Unmarshal(data, &pkg); err != nil {
		return false
	}
	_, hasBuild := pkg.Scripts["build"]
	return hasBuild
}

func (w *BuildWorker) processTask(ctx context.Context, taskID string) {
	var task models.ProjectTask
	if err := w.db.First(&task, "id = ?", taskID).Error; err != nil {
		log.Error().Err(err).Str("task_id", taskID).Msg("failed to fetch task")
		return
	}

	// Update task status to running
	now := time.Now().UTC()
	task.Status = models.TaskStatusRunning
	task.StartedAt = &now
	w.db.Save(&task)

	log.Info().Str("task_id", task.ID).Str("type", string(task.Type)).Str("role", w.Role).Msg("processing task")

	switch w.Role {
	case "syncer":
		w.handleSyncTask(ctx, &task)
	case "builder":
		w.handleBuildTask(ctx, &task)
	case "tester":
		w.handleTestTask(ctx, &task)
	case "deployer":
		w.handleDeployTask(ctx, &task)
	case "ai":
		w.handleAITask(ctx, &task)
	default:
		w.completeTask(task.ID, false, fmt.Sprintf("unsupported role: %s", w.Role))
	}
}

func (w *BuildWorker) handleSyncTask(ctx context.Context, task *models.ProjectTask) {
	project, err := w.getProjectDir(task.ProjectID)
	if err != nil {
		w.completeTask(task.ID, false, fmt.Sprintf("Project not found: %v", err))
		return
	}

	sourcePath := filepath.Join(w.cfg.ProjectsDir, project.ID, task.CommitSHA)
	_ = os.MkdirAll(filepath.Dir(sourcePath), 0755)

	// Mocking job for clone/sync
	job := &models.DeploymentJob{
		ProjectID: project.ID,
		RepoURL:   project.RepoURL,
		Branch:    project.Branch,
		GitToken:  project.GitToken,
		IsPrivate: project.IsPrivate,
	}

	w.appendLog(task.ID, "info", "system", fmt.Sprintf("Cloning repository %s to %s", project.RepoURL, sourcePath))
	if err := w.cloneRepo(ctx, job, sourcePath); err != nil {
		w.appendLog(task.ID, "error", "system", fmt.Sprintf("Clone failed: %v", err))
		w.completeTask(task.ID, false, fmt.Sprintf("clone failed: %v", err))
		return
	}
	w.appendLog(task.ID, "info", "system", "Repository cloned successfully")

	// Capture latest commit metadata and update Project model
	if sha, msg, author, dateStr, err := getRepoCommitInfo(sourcePath); err == nil && sha != "" {
		task.CommitSHA = sha
		w.db.Model(&models.ProjectTask{}).Where("id = ?", task.ID).Update("commit_sha", sha)

		// Parse date (git %ai is usually 2023-05-19 14:30:05 +0530)
		var commitDate *time.Time
		if t, err := time.Parse("2006-01-02 15:04:05 -0700", dateStr); err == nil {
			commitDate = &t
		}

		w.db.Model(&models.Project{}).Where("id = ?", project.ID).Updates(map[string]interface{}{
			"latest_commit_sha":  sha,
			"latest_commit_msg":  msg,
			"latest_commit_at":   commitDate,
			"updated_at":         time.Now().UTC(),
		})
		
		w.appendLog(task.ID, "info", "system", fmt.Sprintf("Updated project metadata: %s by %s", sha[:7], author))
	}

	w.completeTask(task.ID, true, "")
}

func (w *BuildWorker) handleBuildTask(ctx context.Context, task *models.ProjectTask) {
	project, err := w.getProjectDir(task.ProjectID)
	if err != nil {
		w.completeTask(task.ID, false, fmt.Sprintf("Project not found: %v", err))
		return
	}

	job := &models.DeploymentJob{
		DeploymentID: task.ID,
		ProjectID:    task.ProjectID,
		CommitSHA:    task.CommitSHA,
		RepoURL:      project.RepoURL,
		Branch:       project.Branch,
		BuildCommand: project.BuildCommand,
		ImageTag:     fmt.Sprintf("pushpaka-%s:%s", task.ProjectID, task.CommitSHA),
	}

	w.processJob(ctx, job)
	
	// Check if processJob ended in failure
	var updatedTask models.ProjectTask
	w.db.First(&updatedTask, "id = ?", task.ID)
	if updatedTask.Status == models.TaskStatusRunning {
		w.completeTask(task.ID, true, "")
	}
}

func (w *BuildWorker) handleDeployTask(ctx context.Context, task *models.ProjectTask) {
	project, err := w.getProjectDir(task.ProjectID)
	if err != nil {
		w.completeTask(task.ID, false, fmt.Sprintf("Project not found: %v", err))
		return
	}

	// For deploying, we promote BUILDS_DIR artifacts to DEPLOYS_DIR and run.
	buildsDir := filepath.Join(w.cfg.BuildsDir, project.ID, task.CommitSHA)
	deployDir := filepath.Join(w.cfg.DeploysDir, project.ID, task.CommitSHA)

	if _, err := os.Stat(buildsDir); os.IsNotExist(err) {
		w.completeTask(task.ID, false, "build artifacts not found for deployment. Please build first.")
		return
	}

	// Clean/Prepare deployment directory
	_ = os.RemoveAll(deployDir)
	if err := os.MkdirAll(deployDir, 0755); err != nil {
		w.completeTask(task.ID, false, fmt.Sprintf("failed to create deployment directory: %v", err))
		return
	}

	w.appendLog(task.ID, "info", "system", "Promoting build artifacts to deployment directory...")
	if err := copyDir(buildsDir, deployDir); err != nil {
		w.completeTask(task.ID, false, fmt.Sprintf("failed to copy artifacts: %v", err))
		return
	}

	// Find deployment record to get assigned port
	var dep models.Deployment
	if err := w.db.Where("project_id = ? AND commit_sha = ? AND status = ?", task.ProjectID, task.CommitSHA, "queued").First(&dep).Error; err != nil {
		// Fallback if no queued deployment found, maybe it's a direct task
		w.appendLog(task.ID, "info", "system", "No queued deployment record found, using project defaults")
	}

	job := &models.DeploymentJob{
		DeploymentID:    task.ID,
		ProjectID:       task.ProjectID,
		CommitSHA:       task.CommitSHA,
		RepoURL:         project.RepoURL,
		Branch:          project.Branch,
		InstallCommand:  project.InstallCommand,
		BuildCommand:    project.BuildCommand,
		StartCommand:    project.StartCommand,
		RunDir:          project.RunDir,
		Port:            project.Port,
		ExternalPort:    dep.ExternalPort,
		IsRecovery:      true, // Skip build, we already copied artifacts
	}

	if job.ExternalPort == 0 {
		// Assign a port if missing
		if addr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:0"); err == nil {
			if l, err := net.ListenTCP("tcp", addr); err == nil {
				job.ExternalPort = l.Addr().(*net.TCPAddr).Port
				l.Close()
			}
		}
	}

	w.processJob(ctx, job)

	// Check if processJob ended in failure
	var updatedTask models.ProjectTask
	w.db.First(&updatedTask, "id = ?", task.ID)
	if updatedTask.Status == models.TaskStatusRunning {
		w.completeTask(task.ID, true, "")
	}
}

func (w *BuildWorker) handleTestTask(ctx context.Context, task *models.ProjectTask) {
	project, err := w.getProjectDir(task.ProjectID)
	if err != nil {
		w.completeTask(task.ID, false, fmt.Sprintf("Project not found: %v", err))
		return
	}

	// For testing, we copy BUILDS_DIR artifacts to TESTS_DIR and run tests
	buildsDir := filepath.Join(w.cfg.BuildsDir, project.ID, task.CommitSHA)
	testDir := filepath.Join(w.cfg.TestsDir, project.ID, task.CommitSHA)

	if _, err := os.Stat(buildsDir); os.IsNotExist(err) {
		w.completeTask(task.ID, false, "build artifacts not found for testing")
		return
	}

	_ = os.RemoveAll(testDir)
	_ = os.MkdirAll(testDir, 0755)
	if err := copyDir(buildsDir, testDir); err != nil {
		w.completeTask(task.ID, false, fmt.Sprintf("failed to setup test directory: %v", err))
		return
	}

	// 1. Allocate a random port for isolated testing
	testPort := 0
	if addr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:0"); err == nil {
		if l, err := net.ListenTCP("tcp", addr); err == nil {
			testPort = l.Addr().(*net.TCPAddr).Port
			l.Close()
		}
	}
	if testPort == 0 {
		testPort = 13000 + (int(time.Now().UnixNano()) % 5000)
	}

	w.appendLog(task.ID, "info", "system", fmt.Sprintf("Starting isolated test instance on port %d...", testPort))
	w.appendLog(task.ID, "info", "system", fmt.Sprintf("Test Directory: %s", testDir))

	// 2. Start the application in the background
	startCmd := project.StartCommand
	if startCmd == "" {
		w.appendLog(task.ID, "info", "system", "No start command defined, detecting...")
		pm := detectPackageManager(testDir)
		startCmd = detectNodeStartCmd(testDir, pm, testPort)
	}

	w.appendLog(task.ID, "info", "system", fmt.Sprintf("Execution command: %s", startCmd))

	// Replace port placeholder if exists, otherwise set PORT env
	startCmd = strings.ReplaceAll(startCmd, "$PORT", fmt.Sprintf("%d", testPort))
	startCmd = strings.ReplaceAll(startCmd, "{{port}}", fmt.Sprintf("%d", testPort))

	shell, shellFlag := "sh", "-c"
	if runtime.GOOS == "windows" {
		shell, shellFlag = "cmd", "/c"
	}
	proc := exec.CommandContext(ctx, shell, shellFlag, startCmd)
	proc.Dir = testDir
	// Inherit env and add PORT
	proc.Env = append(os.Environ(), fmt.Sprintf("PORT=%d", testPort), "NODE_ENV=test", "APP_ENV=test")
	
	stdoutPipe, _ := proc.StdoutPipe()
	stderrPipe, _ := proc.StderrPipe()
	
	if err := proc.Start(); err != nil {
		w.completeTask(task.ID, false, fmt.Sprintf("failed to start test instance: %v", err))
		return
	}

	// Stream logs in background
	go io.Copy(&logWriter{deploymentID: task.ID, stream: "stdout", w: w}, stdoutPipe)
	go io.Copy(&logWriter{deploymentID: task.ID, stream: "stderr", w: w}, stderrPipe)

	// 3. Wait for app to be ready (health check)
	ready := false
	testURL := fmt.Sprintf("http://127.0.0.1:%d", testPort)
	for i := 0; i < 15; i++ {
		if w.healthCheck(ctx, testURL) {
			ready = true
			break
		}
		time.Sleep(1 * time.Second)
	}

	if !ready {
		_ = proc.Process.Kill()
		w.completeTask(task.ID, false, "application failed to start or pass health check on test port")
		return
	}

	// 4. Run the actual test command
	testCmd := project.TestCommand
	if testCmd != "" {
		w.appendLog(task.ID, "info", "system", fmt.Sprintf("Running test command: %s", testCmd))
		tCmd := exec.CommandContext(ctx, shell, shellFlag, testCmd)
		tCmd.Dir = testDir
		tCmd.Env = append(os.Environ(), fmt.Sprintf("TEST_URL=%s", testURL), fmt.Sprintf("PORT=%d", testPort))
		tCmd.Stdout = &logWriter{deploymentID: task.ID, stream: "stdout", w: w}
		tCmd.Stderr = &logWriter{deploymentID: task.ID, stream: "stderr", w: w}
		
		if err := tCmd.Run(); err != nil {
			_ = proc.Process.Kill()
			w.completeTask(task.ID, false, fmt.Sprintf("Test command failed: %v", err))
			return
		}
	} else {
		w.appendLog(task.ID, "info", "system", "No test command defined, health check passed.")
	}

	// 5. Cleanup
	_ = proc.Process.Kill()
	w.appendLog(task.ID, "info", "system", "Test instance stopped. Cleanup complete.")
	w.completeTask(task.ID, true, "")
}

func (w *BuildWorker) getProjectDir(id string) (*models.Project, error) {
	var p models.Project
	if err := w.db.First(&p, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func (w *BuildWorker) completeTask(id string, success bool, errStr string) {
	// 1. Update DB directly (for speed/fallback)
	status := models.TaskStatusCompleted
	if !success {
		status = models.TaskStatusFailed
	}
	w.db.Model(&models.ProjectTask{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":      status,
		"error":       errStr,
		"finished_at": time.Now().UTC(),
	})

	// 2. Notify backend to trigger next task
	apiURL := fmt.Sprintf("%s/api/v1/internal/tasks/%s/complete", w.cfg.ServerURL, id)
	
	payload, _ := json.Marshal(map[string]interface{}{
		"success": success,
		"error":   errStr,
	})
	resp, err := http.Post(apiURL, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		log.Error().Err(err).Str("task_id", id).Msg("failed to notify backend of task completion")
		return
	}
	defer resp.Body.Close()
}
func (w *BuildWorker) handleAITask(ctx context.Context, task *models.ProjectTask) {
	w.appendLog(task.ID, "info", "system", "AI Assistant processing task...")
	
	if w.cfg.AIAPIKey == "" {
		w.completeTask(task.ID, false, "AI API key not configured")
		return
	}

	prompt := fmt.Sprintf("Reviewing project task: %s\nSummary: %s", task.Type, task.Log)
	w.appendLog(task.ID, "info", "system", "Analyzing task with AI: " + prompt)
	
	// Just a simple placeholder for now that uses analyzeFailure if it was an error
	// or provide a generic success summary
	summary := "AI task processed successfully."
	if task.Error != "" {
		summary = w.analyzeFailure(task.ID, task.Error)
	}
	
	w.appendLog(task.ID, "info", "system", "AI Analysis Results: " + summary)
	w.completeTask(task.ID, true, "")
}
