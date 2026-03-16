You are a senior platform engineer and product architect.

Design and implement a **production-grade self-hosted cloud deployment platform** named **Pushpaka**.

Pushpaka should function similarly to platforms like **Vercel, Render, and Railway**, allowing developers to deploy applications directly from Git repositories.

The system must be fully functional, production-ready, modular, and follow clean architecture principles.

---

# Product Identity

Application Name:
Pushpaka

Version:
v1.0.0

Tagline:
“Pushpaka — Carry your code to the cloud effortlessly.”

Description:
Pushpaka is a modern self-hosted cloud platform that allows developers to deploy applications directly from Git repositories with automated builds, containerized deployment, custom domains, and scalable infrastructure.

Pushpaka should feel like a **real SaaS developer platform**.

---

# Branding

Create branding assets for the platform.

Generate:

1. A modern **Pushpaka logo**
2. SVG logo file
3. Favicon
4. OpenGraph preview image
5. Tailwind-compatible brand color palette
6. Logo stored under `/branding`

Logo concept:
A futuristic **flying cloud platform inspired by Pushpaka Vimana**, symbolizing carrying applications to the cloud.

Style:
modern
minimal
developer-focused
cloud infrastructure aesthetic

---

# System Architecture

Pushpaka must be a **full-stack distributed platform**.

Structure the project as a **monorepo**.

```
pushpaka/
  frontend/
  backend/
  worker/
  infrastructure/
  branding/
  docs/
```

---

# Frontend

Framework:
Next.js (latest stable)

UI:
TailwindCSS

Features:

Developer Dashboard

Login/Register

Project creation

Git repository connection

Deployment history

Real-time deployment logs

Environment variable manager

Custom domain manager

Settings panel

Dark/light mode

Deployment status indicators

Live log streaming using WebSockets.

Pages:

Dashboard

Create Project

Deployment Logs

Environment Variables

Domains

Account Settings

---

# Backend

Language:
Go

Framework:
Gin or Fiber

Responsibilities:

API server

Deployment orchestration

Authentication

Project management

Deployment tracking

Log aggregation

Domain configuration

Use **clean architecture**:

```
internal/
  handlers/
  services/
  repositories/
  models/
  middleware/
```

---

# Worker System

Implement background workers responsible for:

Repository cloning

Build execution

Docker image creation

Container deployment

Rollback

Workers should process jobs from a queue.

Queue system:

Redis queue

Worker service:

```
worker/
  build_worker.go
  deploy_worker.go
```

---

# Infrastructure

Use containerized infrastructure.

Components:

Docker

Traefik reverse proxy

PostgreSQL database

Redis queue

Container runtime for deployments

Services:

pushpaka-api

pushpaka-worker

pushpaka-dashboard

postgres

redis

traefik

Provide a **complete docker-compose setup**.

---

# Deployment Flow

The deployment pipeline should work as follows:

1. User connects Git repository
2. Push event triggers deployment
3. Repository cloned
4. Build executed
5. Docker image created
6. Container deployed
7. Traefik routes traffic
8. Logs streamed to dashboard

---

# Core Platform Features

GitHub repository integration

Automatic deployment on push

Docker image build

Container runtime deployment

Reverse proxy routing via Traefik

Custom domain support

Environment variable configuration

Deployment logs with live streaming

Rollback to previous deployments

Multi-project support

Multi-user support

JWT authentication

Rate limiting

Audit logging

Health monitoring

---

# API Endpoints

Auth:

POST /auth/register
POST /auth/login

Projects:

POST /projects
GET /projects

Deployments:

POST /deploy
GET /deployments
GET /deployments/{id}

Logs:

GET /logs/{deployment}

Domains:

POST /domains
GET /domains

Environment Variables:

POST /env
GET /env

---

# Database

PostgreSQL schema must include:

users

projects

deployments

domains

environment_variables

deployment_logs

---

# Security

Implement production-grade security:

JWT authentication

API key system

password hashing (bcrypt)

rate limiting

secure headers

CORS configuration

---

# Observability

Add observability features:

structured logging

deployment metrics

Prometheus metrics endpoint

health checks

---

# Documentation

Generate full documentation inside `/docs`.

Include:

architecture diagram

deployment guide

local development guide

API documentation

platform overview

---

# Deliverables

Generate a **fully working project** including:

Complete backend source code

Complete frontend dashboard

Worker system

Dockerfiles

docker-compose.yml

Database migrations

Branding assets and logo

Example configuration files

Production-ready folder structure

Sample environment variables

Seed data for demo projects

---

# README

Generate a professional README including:

Pushpaka logo

Tagline

Architecture diagram

Screenshots

Quick start guide

Docker deployment instructions

Feature list

Roadmap for v1.1.0

---

The final system must be a **fully functional self-hosted developer platform** capable of deploying applications similarly to Vercel or Render.

All code, configuration, infrastructure, and branding should be generated automatically as part of the Pushpaka v1.0.0 release.
