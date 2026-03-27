export type SupportEntry = {
  title: string;
  keywords: string[];
  response: string;
  links?: Array<{ label: string; href: string }>;
};

export const supportQuickPrompts = [
  "How do I install Pushpaka?",
  "How do AI monitoring and log analysis work?",
  "How do I deploy with Docker or Kubernetes?",
  "How do I use the built-in editor and terminal?",
  "How do I upgrade Pushpaka safely?",
];

export const supportEntries: SupportEntry[] = [
  {
    title: "Installation",
    keywords: ["install", "setup", "docker compose", "helm", "start", "run", "quick start"],
    response:
      "Pushpaka supports Docker Compose, Helm, and source-based local development. For the fastest production-style setup, copy .env.example, configure domain and secrets, then run docker compose up -d --build. For Kubernetes, add the Helm repo from pushpaka.vikshro.in/helm and install the pushpaka chart.",
    links: [
      { label: "Installation Guide", href: "/install" },
      { label: "Helm Install", href: "/helm-install" },
      { label: "Docs", href: "/docs" },
    ],
  },
  {
    title: "AI Operations",
    keywords: ["ai", "monitoring", "alerts", "log", "analysis", "assistant", "rag", "observability"],
    response:
      "Pushpaka v1.0.0 includes AI-assisted operations for deployment log analysis, monitoring workflows, assistant-style troubleshooting, and runbook context. Typical usage is: open a failed deployment, inspect realtime logs, run AI analysis, review remediation hints, then redeploy or roll back.",
    links: [
      { label: "Feature Overview", href: "/features" },
      { label: "Operations Docs", href: "/docs#ai-operations" },
    ],
  },
  {
    title: "Deployment Targets",
    keywords: ["deploy", "docker", "kubernetes", "k8s", "runtime", "worker", "platform", "branch"],
    response:
      "Pushpaka handles Git-based deployments across Docker workflows, Kubernetes-oriented targets, and direct runtime fallback. Teams can run single-binary dev mode, all-in-one production mode, or split API and worker mode with Redis-backed queueing.",
    links: [
      { label: "Runtime Modes", href: "/docs#deployment-modes" },
      { label: "Install", href: "/install" },
    ],
  },
  {
    title: "Editor And Terminal",
    keywords: ["editor", "terminal", "monaco", "files", "source", "workspace", "shell", "debug"],
    response:
      "Pushpaka includes a built-in Monaco editor and web terminal so operators can inspect, edit, sync, and validate application files without leaving the platform. A common workflow is to open the project editor, sync source, apply a change, validate in terminal, then trigger a redeploy.",
    links: [
      { label: "Operator Docs", href: "/docs#editor-terminal" },
      { label: "Features", href: "/features" },
    ],
  },
  {
    title: "Domains And Routing",
    keywords: ["domain", "dns", "tls", "traefik", "ingress", "ssl", "routing", "certificate"],
    response:
      "Pushpaka uses Traefik-based routing with custom domains and automatic TLS provisioning. Add a project domain, point DNS to the Pushpaka entrypoint, wait for provisioning, and validate the certificate and route from the dashboard or docs workflow.",
    links: [
      { label: "Routing Docs", href: "/docs#domains-routing" },
      { label: "Helm Install", href: "/helm-install" },
    ],
  },
  {
    title: "API And Integrations",
    keywords: ["api", "webhook", "github", "gitlab", "slack", "discord", "smtp", "notification"],
    response:
      "Pushpaka exposes API endpoints for auth, projects, deployments, logs, AI operations, notifications, webhooks, infrastructure controls, and health checks. It also supports GitHub and GitLab OAuth, incoming webhooks, and Slack, Discord, and SMTP notifications.",
    links: [
      { label: "Docs", href: "/docs#api-usage" },
      { label: "Repository", href: "https://github.com/vikukumar/pushpaka" },
    ],
  },
  {
    title: "Upgrade And Release",
    keywords: ["upgrade", "release", "version", "download", "binary", "latest", "release assets"],
    response:
      "Use the Releases page for the latest GitHub release assets and checksums. For packaged deployments, prefer the combined release archives or Docker image. For Helm users, update the repo and upgrade the chart. Always review release notes and verify checksums before rollout.",
    links: [
      { label: "Releases", href: "/releases" },
      { label: "Upgrade Docs", href: "/docs#upgrade-runbook" },
    ],
  },
  {
    title: "Troubleshooting",
    keywords: ["error", "issue", "failed", "not working", "404", "broken", "problem", "troubleshoot"],
    response:
      "Start with deployment logs, health endpoints, and system status. For failed releases, verify the correct asset name and tag. For routing issues, confirm DNS and TLS state. For build problems, inspect logs, AI analysis output, and configuration such as runtime, build command, and environment variables.",
    links: [
      { label: "Troubleshooting Docs", href: "/docs#common-errors" },
      { label: "Support", href: "/support" },
    ],
  },
];

