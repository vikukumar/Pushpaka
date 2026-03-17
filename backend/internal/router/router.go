package router

import (
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/http/httputil"
	"net/url"
	"path"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"

	"github.com/vikukumar/Pushpaka/internal/config"
	"github.com/vikukumar/Pushpaka/internal/handlers"
	"github.com/vikukumar/Pushpaka/internal/middleware"
	"github.com/vikukumar/Pushpaka/internal/repositories"
	"github.com/vikukumar/Pushpaka/internal/services"
	"github.com/vikukumar/Pushpaka/queue"
)

// New builds the Gin engine. Pass a non-nil uiFS to serve the embedded
// Next.js static export under /; pass nil to skip frontend serving (dev mode).
// inQueue is only non-nil in dev mode (embedded worker, no Redis).
func New(cfg *config.Config, db *sqlx.DB, rdb *redis.Client, uiFS fs.FS, inQueue *queue.InProcess) *gin.Engine {
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()

	// Global middleware
	r.Use(middleware.Recovery())
	r.Use(middleware.Logger())
	r.Use(middleware.SecureHeaders())
	r.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.AllowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-API-Key"},
		AllowCredentials: true,
	}))

	// Repositories
	userRepo := repositories.NewUserRepository(db)
	projectRepo := repositories.NewProjectRepository(db)
	deploymentRepo := repositories.NewDeploymentRepository(db)
	logRepo := repositories.NewLogRepository(db)
	domainRepo := repositories.NewDomainRepository(db)
	envRepo := repositories.NewEnvVarRepository(db)
	auditRepo := repositories.NewAuditRepository(db)
	notifRepo := repositories.NewNotificationRepository(db)
	webhookRepo := repositories.NewWebhookRepository(db)
	aiConfigRepo := repositories.NewAIConfigRepository(db)

	// Services
	authSvc := services.NewAuthService(userRepo, cfg)
	projectSvc := services.NewProjectService(projectRepo)
	deploymentSvc := services.NewDeploymentService(deploymentRepo, projectRepo, envRepo, domainRepo, rdb, inQueue, cfg.BaseURL)
	logSvc := services.NewLogService(logRepo)
	domainSvc := services.NewDomainService(domainRepo, projectRepo)
	envSvc := services.NewEnvService(envRepo, projectRepo)
	auditSvc := services.NewAuditService(auditRepo)
	notifSvc := services.NewNotificationService(notifRepo, cfg)
	oauthSvc := services.NewOAuthService(userRepo, cfg, authSvc, db)
	webhookSvc := services.NewWebhookService(webhookRepo, projectRepo, deploymentSvc, cfg)
	aiSvc := services.NewAIService(cfg)

	// Handlers
	authHandler := handlers.NewAuthHandler(authSvc)
	projectHandler := handlers.NewProjectHandler(projectSvc, auditSvc)
	deploymentHandler := handlers.NewDeploymentHandler(deploymentSvc, auditSvc)
	logHandler := handlers.NewLogHandler(logSvc)
	domainHandler := handlers.NewDomainHandler(domainSvc)
	envHandler := handlers.NewEnvHandler(envSvc)
	auditHandler := handlers.NewAuditHandler(auditSvc)
	notifHandler := handlers.NewNotificationHandler(notifSvc)
	oauthHandler := handlers.NewOAuthHandler(oauthSvc)
	webhookHandler := handlers.NewWebhookHandler(webhookSvc)
	aiHandler := handlers.NewAIHandler(aiSvc, logRepo, deploymentRepo, aiConfigRepo, cfg)
	terminalHandler := handlers.NewTerminalHandler(deploymentRepo, cfg)
	fileHandler := handlers.NewFileHandler(projectRepo, deploymentRepo, cfg)
	infraHandler := handlers.NewInfraHandler()
	// Avoid the nil-interface trap: a nil *queue.InProcess assigned directly to
	// WorkerStatsProvider creates a non-nil interface with a nil concrete value,
	// which passes the != nil check but panics on method calls.
	var workerStats handlers.WorkerStatsProvider
	if inQueue != nil {
		workerStats = inQueue
	}
	healthHandler := handlers.NewHealthHandler(db, rdb, workerStats)

	// Auth middleware
	authMW := middleware.JWT(authSvc)

	// API v1
	api := r.Group("/api/v1")
	{
		// Health & Metrics (public)
		api.GET("/health", healthHandler.Health)
		api.GET("/ready", healthHandler.Ready)
		api.GET("/system", healthHandler.System)
		api.GET("/metrics", handlers.MetricsHandler())

		// Auth (public) — rate-limited to prevent brute-force
		auth := api.Group("/auth")
		auth.Use(middleware.RateLimit("auth"))
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			// OAuth flows (public — redirects handle the JWT issuance)
			auth.GET("/github", oauthHandler.GithubRedirect)
			auth.GET("/github/callback", oauthHandler.GithubCallback)
			auth.GET("/gitlab", oauthHandler.GitlabRedirect)
			auth.GET("/gitlab/callback", oauthHandler.GitlabCallback)
		}

		// Incoming webhooks (public — signature-verified)
		api.POST("/webhooks/:id/receive", webhookHandler.Receive)

		// Internal notification callback (only called by the worker, not exposed publicly)
		api.POST("/internal/notify", notifHandler.InternalNotify)

		// Protected routes
		protected := api.Group("")
		protected.Use(authMW)
		protected.Use(middleware.RateLimit("api"))
		{
			// Projects
			projects := protected.Group("/projects")
			{
				projects.POST("", projectHandler.Create)
				projects.GET("", projectHandler.List)
				projects.GET("/:id", projectHandler.Get)
				projects.PUT("/:id", projectHandler.Update)
				projects.DELETE("/:id", projectHandler.Delete)
			}

			// Deployments
			deployments := protected.Group("/deployments")
			{
				deployments.POST("", deploymentHandler.Deploy)
				deployments.GET("", deploymentHandler.List)
				deployments.GET("/:id", deploymentHandler.Get)
				deployments.POST("/:id/rollback", deploymentHandler.Rollback)
			}

			// Logs (REST + WebSocket)
			logs := protected.Group("/logs")
			{
				logs.GET("/:id", logHandler.GetLogs)
				logs.GET("/:id/stream", logHandler.StreamLogs)
			}

			// Domains
			domains := protected.Group("/domains")
			{
				domains.POST("", domainHandler.Add)
				domains.GET("", domainHandler.List)
				domains.DELETE("/:id", domainHandler.Delete)
			}

			// Environment Variables
			env := protected.Group("/env")
			{
				env.POST("", envHandler.Set)
				env.GET("", envHandler.List)
				env.DELETE("", envHandler.Delete)
			}

			// Webhooks (management; the receive endpoint is public above)
			webhooks := protected.Group("/webhooks")
			{
				webhooks.POST("", webhookHandler.Create)
				webhooks.GET("", webhookHandler.List)
				webhooks.DELETE("/:id", webhookHandler.Delete)
			}

			// Notifications config
			notifications := protected.Group("/notifications")
			{
				notifications.GET("/config", notifHandler.Get)
				notifications.PUT("/config", notifHandler.Upsert)
			}

			// Audit logs
			protected.GET("/audit", auditHandler.List)

			// Web terminal (WebSocket)
			protected.GET("/deployments/:id/terminal", terminalHandler.Connect)

			// AI log analysis + chat assistant + config
			protected.POST("/deployments/:id/analyze", aiHandler.AnalyzeLogs)
			protected.POST("/ai/chat", aiHandler.Chat)
			protected.GET("/ai/config", aiHandler.GetAIConfig)
			protected.PUT("/ai/config", aiHandler.SaveAIConfig)

			// RAG knowledge base
			protected.GET("/ai/rag", aiHandler.ListRAG)
			protected.POST("/ai/rag", aiHandler.CreateRAG)
			protected.DELETE("/ai/rag/:id", aiHandler.DeleteRAG)

			// AI monitoring alerts
			protected.GET("/ai/alerts", aiHandler.ListAlerts)
			protected.PUT("/ai/alerts/:id/resolve", aiHandler.ResolveAlert)

			// AI usage stats (how many calls used today vs limit)
			protected.GET("/ai/usage", aiHandler.GetUsage)

			// Docker container management
			protected.GET("/containers", infraHandler.ListContainers)
			protected.POST("/containers/:id/start", infraHandler.StartContainer)
			protected.POST("/containers/:id/stop", infraHandler.StopContainer)
			protected.POST("/containers/:id/restart", infraHandler.RestartContainer)
			protected.GET("/containers/:id/logs", infraHandler.ContainerLogs)

			// Kubernetes management
			protected.GET("/k8s/namespaces", infraHandler.K8sNamespaces)
			protected.GET("/k8s/pods", infraHandler.K8sPods)
			protected.GET("/k8s/deployments", infraHandler.K8sDeployments)
			protected.GET("/k8s/services", infraHandler.K8sServices)
			protected.POST("/k8s/deployments/:namespace/:name/rollout", infraHandler.K8sRollout)

			// Deployment delete
			protected.DELETE("/deployments/:id", deploymentHandler.Delete)

			// In-browser code editor (file browser + read + save)
			protected.GET("/projects/:id/files", fileHandler.ListFiles)
			protected.GET("/projects/:id/files/*path", fileHandler.ReadFile)
			protected.PUT("/projects/:id/files/*path", fileHandler.SaveFile)
			// Sync (re-clone) project source to the editor working directory
			protected.POST("/projects/:id/files/sync", fileHandler.SyncFiles)
		}
	}

	// /app/:projectID -- path-based deployment access (when no custom domain is set).
	// Proxies all requests to the running container's external port.
	r.Any("/app/:projectID/*proxyPath", newDeploymentProxyHandler(deploymentRepo))
	r.Any("/app/:projectID", newDeploymentProxyHandler(deploymentRepo))

	// Serve embedded frontend SPA (only when built with frontend assets).
	if uiFS != nil {
		spaH := newSPAHandler(uiFS)
		r.NoRoute(func(c *gin.Context) {
			spaH.ServeHTTP(c.Writer, c.Request)
		})
	}

	return r
}

// newSPAHandler returns an http.Handler that serves a Next.js static export.
//
// Next.js App Router makes two kinds of requests:
//   - Full page / hard navigation  -> GET /dashboard/       (no _rsc param) -> serve .html
//   - RSC navigation fetch         -> GET /dashboard/?_rsc= -> serve .txt with text/x-component
//
// Resolution order for both kinds:
//  1. Exact non-directory file (JS, CSS, images, fonts, explicit .html ...)
//  2. RSC request -> path/index.txt  (RSC payload pre-built by Next.js static export)
//  3. HTML request -> path/index.html (full pre-rendered page)
//  4. Dynamic-segment placeholder: substitute unknown UUID segments with "_"
//  5. SPA fallback -> root index.html
func newSPAHandler(fsys fs.FS) http.Handler {
	fsh := http.FileServerFS(fsys)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upath := r.URL.Path
		if !strings.HasPrefix(upath, "/") {
			upath = "/" + upath
		}
		upath = path.Clean(upath)
		name := strings.TrimPrefix(upath, "/")
		if name == "" {
			name = "index.html"
		}

		// 1. Exact non-directory file (CSS, JS, images, explicit .html, etc.)
		if fi, err := fs.Stat(fsys, name); err == nil && !fi.IsDir() {
			fsh.ServeHTTP(w, r)
			return
		}

		// 2 & 3. Handle RSC navigation vs full-page requests differently.
		//
		// Next.js App Router adds ?_rsc=<ts> to its navigation fetch and sets
		// Accept: text/x-component.  The pre-built static export stores the RSC
		// payload alongside the HTML as  <route>/index.txt.  Returning HTML for
		// this fetch causes the RSC parser to throw "Uncaught (in promise) Event"
		// and navigation silently fails.
		baseName := strings.TrimRight(name, "/")
		isRSC := r.URL.Query().Has("_rsc") ||
			strings.Contains(r.Header.Get("Accept"), "text/x-component")

		if isRSC {
			// Serve RSC payload with the correct content-type.
			w.Header().Set("Content-Type", "text/x-component; charset=utf-8")

			// The path may contain an explicit /index.txt suffix (Next.js 16 Turbopack
			// requests /dashboard/projects/<uuid>/index.txt?_rsc=...).  Strip it so we
			// work with the bare directory path in all steps below.
			dirBase := strings.TrimRight(name, "/")
			dirBase = strings.TrimSuffix(dirBase, "/index.txt")

			// 2a. Exact index.txt for this directory.
			if serveFile(w, r, fsys, dirBase+"/index.txt") {
				return
			}
			// 2b. Dynamic-segment fallback: /projects/<uuid>/ -> projects/_/index.txt
			if segs := splitSegments(dirBase); len(segs) > 0 {
				if resolved := resolveFile(fsys, segs, "index.txt"); resolved != "" {
					serveFile(w, r, fsys, resolved)
					return
				}
			}

			// 2c. __next.X.Y.Z.txt segment payload files.
			//
			// Next.js Turbopack stores segment RSC payloads in a directory tree:
			//   Request filename:  __next.dashboard.projects.$d$id.deployments.__PAGE__.txt
			//   Converted path:    __next.dashboard/projects/$d$id/deployments/__PAGE__.txt
			//   Actual file (FS):  dashboard/projects/_/deployments/__next.dashboard/projects/$d$id/deployments/__PAGE__.txt
			//
			// Two transforms are needed:
			//   1. Dot-to-slash on the filename  ("X.Y.Z" -> "X/Y/Z")
			//   2. UUID-to-placeholder on parent dir segments
			base := path.Base(name)
			if strings.HasPrefix(base, "__next.") && strings.HasSuffix(base, ".txt") {
				dir := path.Dir(name)
				dirSegs := splitSegments(dir)
				inner := strings.TrimPrefix(strings.TrimSuffix(base, ".txt"), "__next.")
				if dot := strings.Index(inner, "."); dot != -1 {
					firstSeg := inner[:dot]
					rest := strings.ReplaceAll(inner[dot+1:], ".", "/")
					convertedFile := "__next." + firstSeg + "/" + rest + ".txt"
					// 2c-i: exact (no dynamic segments in parent path)
					dirPrefix := ""
					if dir != "." {
						dirPrefix = dir + "/"
					}
					if serveFile(w, r, fsys, dirPrefix+convertedFile) {
						return
					}
					// 2c-ii: UUID placeholder substitution in parent dir segments
					if len(dirSegs) > 0 {
						if resolved := resolveFile(fsys, dirSegs, convertedFile); resolved != "" {
							serveFile(w, r, fsys, resolved)
							return
						}
					}
				}
			}

			// 2d. Other RSC files (e.g. __next._tree.txt, __next._head.txt) nested
			// inside a dynamic path that contains UUID segments.
			// e.g.: dashboard/projects/<uuid>/deployments/__next._tree.txt
			//   ->   dashboard/projects/_/deployments/__next._tree.txt
			{
				fileName := path.Base(dirBase)
				parentSegs := splitSegments(path.Dir(dirBase))
				if len(parentSegs) > 0 {
					if resolved := resolveFile(fsys, parentSegs, fileName); resolved != "" {
						serveFile(w, r, fsys, resolved)
						return
					}
				}
			}

			http.NotFound(w, r)
			return
		}

		// 3. Full-page (hard navigation): serve the pre-rendered HTML.
		// Directory-style URL: /dashboard/ -> dashboard/index.html
		if serveFile(w, r, fsys, baseName+"/index.html") {
			return
		}

		// 4. File-style URL: /dashboard -> dashboard.html (trailingSlash: false)
		if serveFile(w, r, fsys, baseName+".html") {
			return
		}

		// 5. Dynamic-segment resolution: /projects/<uuid>/ -> projects/_/index.html
		if segs := splitSegments(baseName); len(segs) > 0 {
			if resolved := resolveFile(fsys, segs, "index.html"); resolved != "" {
				serveFile(w, r, fsys, resolved)
				return
			}
		}

		// 6. SPA fallback -> root index.html
		serveFile(w, r, fsys, "index.html")
	})
}

// serveFile serves a specific file from the FS using http.ServeContent,
// bypassing FileServer's index.html->directory redirect behaviour.
// Returns true if the file was found and served.
func serveFile(w http.ResponseWriter, r *http.Request, fsys fs.FS, name string) bool {
	f, err := fsys.Open(name)
	if err != nil {
		return false
	}
	defer f.Close()
	fi, err := f.Stat()
	if err != nil || fi.IsDir() {
		return false
	}
	http.ServeContent(w, r, fi.Name(), fi.ModTime(), f.(io.ReadSeeker))
	return true
}

// splitSegments splits a clean URL path (no leading slash) into non-empty segments.
func splitSegments(name string) []string {
	var out []string
	for _, s := range strings.Split(name, "/") {
		if s != "" {
			out = append(out, s)
		}
	}
	return out
}

// resolveFile finds the best static file for URL segments by substituting
// unknown path segments with the "_" placeholder produced by generateStaticParams
// in the Next.js static export.  file is "index.html" or "index.txt".
//
// Example: (["dashboard","projects","abc-123","env"], "index.html")
//
//	-> "dashboard/projects/_/env/index.html"
func resolveFile(fsys fs.FS, segments []string, file string) string {
	var try func(idx int, prefix string) string
	try = func(idx int, prefix string) string {
		if idx == len(segments) {
			candidate := prefix + file
			if _, err := fs.Stat(fsys, candidate); err == nil {
				return candidate
			}
			return ""
		}
		seg := segments[idx]
		if result := try(idx+1, prefix+seg+"/"); result != "" {
			return result
		}
		if seg != "_" {
			if result := try(idx+1, prefix+"_/"); result != "" {
				return result
			}
		}
		return ""
	}
	return try(0, "")
}

// newDeploymentProxyHandler returns a gin.HandlerFunc that reverse-proxies
// requests to the running container for the given :projectID.
// The container must have a non-zero ExternalPort set by the deployment worker.
func newDeploymentProxyHandler(repo *repositories.DeploymentRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		projectID := c.Param("projectID")

		d, err := repo.FindRunningByProjectID(projectID)
		if err != nil || d == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "no running deployment for this project"})
			return
		}
		if d.ExternalPort == 0 {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "deployment has no exposed port yet"})
			return
		}

		target, _ := url.Parse(fmt.Sprintf("http://localhost:%d", d.ExternalPort))
		proxy := httputil.NewSingleHostReverseProxy(target)

		// Strip the /app/:projectID prefix so the downstream app sees a clean path.
		proxyPath := c.Param("proxyPath")
		if proxyPath == "" {
			proxyPath = "/"
		}
		c.Request.URL.Path = proxyPath
		c.Request.URL.RawPath = proxyPath
		c.Request.Host = target.Host

		proxy.ServeHTTP(c.Writer, c.Request)
	}
}
