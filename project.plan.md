# SUMERU MASTER PLAN

## Modular AI-Native ERP Framework in Go

Version: 1.0  
Architecture Style: Modular + XML Declarative + AI Native  
Core Language: Go (Golang)

---

# 1. VISION

The goal is to build a:

```text id="t5w5zd"
Modern ERP Platform
+
Modular addon architecture
+
XML declarative UI
+
High-performance Go backend
+
AI-native workflow system
+
Plug-and-play addon ecosystem
```

The platform should provide:

- enterprise-grade ERP capabilities
- developer-friendly architecture
- plug-and-play module installation
- XML-driven dynamic UI
- scalable backend
- AI-assisted business automation
- zero dependency on React/Vue
- fully extensible ecosystem

---

# 2. CORE PHILOSOPHY

The system must follow these principles:

---

## 2.1 Modular by Default

Every feature must be an addon/module.

Examples:

- CRM
- Sales
- Inventory
- Accounting
- HR
- Payroll
- Manufacturing

Nothing should be hardcoded into the core.

---

## 2.2 Plug-and-Play Addons

The ERP must support:

```text id="8n4iwv"
Install Addon
→ Register Models
→ Register Views
→ Register Actions
→ Register Security
→ Activate Automatically
```

Adding modules should be as simple as:

```bash id="26dgf3"
sumeru addon install sales
```

OR

```bash id="ed9x7f"
git clone addon
```

Then restart server.

The system automatically:

- discovers addon
- registers models
- imports XML
- applies migrations
- activates routes
- loads menus

The same install → register → XML → menus flow applies.

---

## 2.3 Declarative UI

UI must be defined using XML.

Example:

```xml id="jlwm8x"
<form model="sale.order">
    <field name="customer_id"/>
    <field name="amount_total"/>
</form>
```

The frontend runtime dynamically renders UI.

---

## 2.4 Backend-First Architecture

The backend is the source of truth.

The system should prioritize:

- strong typing
- scalability
- transactional consistency
- concurrency
- security

---

## 2.5 AI-Native ERP

AI should not be an afterthought.

AI must become:

- assistant
- workflow participant
- analytics engine
- automation operator

---

# 3. HIGH-LEVEL SYSTEM ARCHITECTURE

```text id="qmbmjq"
Browser
   ↓
Vanilla JS Runtime
   ↓
View Engine
   ↓
RPC Layer
   ↓
Action Engine
   ↓
ORM Engine
   ↓
Security Engine
   ↓
Workflow Engine
   ↓
PostgreSQL
```

---

# 4. CORE SYSTEMS

The platform contains 10 major engines.

---

# 4.1 ORM ENGINE

## Purpose

Provide:

- models
- fields
- relations
- computed fields
- constraints
- search domains
- transactions

---

## Features

### Models

```go id="e72fqx"
type SaleOrder struct {
    CustomerID int
    Amount     float64
}
```

---

### Relations

Support:

- many2one
- one2many
- many2many

---

### Computed Fields

Example:

```go id="ccwd1d"
Total = Qty * Price
```

---

### Lifecycle Hooks

Support:

- before_create
- after_create
- before_write
- after_write
- before_delete

---

### Search Domains

Domain tuple filtering:

```text id="5b9e4x"
[("state", "=", "confirmed")]
```

---

### Transactions

Every action should support rollback.

---

# 4.2 MODULE ENGINE

This is one of the most important systems.

---

# PLUG-AND-PLAY ADDON SYSTEM

## Vision

Anyone should be able to:

- create addon
- install addon
- share addon
- activate addon
- remove addon

WITHOUT modifying core code.

---

# Addon Structure

```text id="6j8cr8"
addons/
└── sales/
    ├── manifest.json
    ├── models.go
    ├── views.xml
    ├── actions.xml
    ├── menus.xml
    ├── security.xml
    ├── routes.go
    ├── migrations/
    └── static/
```

---

# manifest.json

```json id="1q1azd"
{
  "name": "sales",
  "version": "1.0",
  "depends": ["base", "contacts"],
  "author": "Sumeru",
  "description": "Sales Management Module"
}
```

---

# Module Loader Responsibilities

The loader must:

## 1. Scan addons directory

```text id="5vok9x"
addons/*
```

---

## 2. Read manifests

Extract:

- dependencies
- metadata
- version

---

## 3. Resolve dependency graph

Example:

```text id="5m13v7"
sales
 └── contacts
      └── base
```

---

## 4. Register models

Example:

```go id="fop6j2"
orm.RegisterModel(SaleOrder)
```

---

## 5. Import XML

Load:

- views
- menus
- actions
- security

---

## 6. Execute migrations

Run SQL migrations automatically.

---

## 7. Activate routes

Register:

- HTTP routes
- RPC handlers
- APIs

---

## 8. Register assets

Load:

- CSS
- JS
- templates

---

# Hot Reloading

Future goal:

```text id="efzjvt"
Install Addon
→ Reload Registry
→ Add Menus
→ Activate Instantly
```

Without server restart.

---

# Addon Marketplace Vision

Eventually support:

- addon registry
- community marketplace
- versioning
- dependency resolution

Like:

- App stores (e.g. marketplace UIs)
- npm
- pip

---

# 4.3 XML VIEW ENGINE

## Purpose

Dynamically render UI from XML.

---

# Supported Views

## Form View

```xml id="0l6bf3"
<form>
    <field name="name"/>
</form>
```

---

## Tree View

```xml id="b29fpp"
<tree>
    <field name="name"/>
</tree>
```

---

## Kanban View

```xml id="oqkly4"
<kanban>
    <templates>
        <t>
            <div>{{name}}</div>
        </t>
    </templates>
</kanban>
```

---

## Search View

Support:

- filters
- domains
- grouping

---

## Dashboard View

Support:

- charts
- KPIs
- metrics

---

# UI Runtime Responsibilities

The runtime should:

- parse XML
- generate HTML
- bind actions
- update DOM dynamically

---

# 4.4 ACTION ENGINE

## Purpose

Allow UI buttons to execute backend methods.

---

# Example

```xml id="l5v0c4"
<button name="confirm_order" type="object"/>
```

---

# Action Types

## Object Actions

Execute model methods.

---

## Server Actions

Run backend logic.

---

## Window Actions

Open views/pages.

---

## Scheduled Actions

Cron jobs.

---

## Workflow Actions

State transitions.

---

# 4.5 SECURITY ENGINE

Security is mandatory.

---

# Features

## Authentication

- sessions
- JWT
- OAuth optional

---

## ACL

Control CRUD permissions.

---

## Groups

Examples:

- admin
- manager
- employee

---

## Record Rules

Restrict row-level access.

Example:

```text id="65q0qq"
Users only see their own sales orders.
```

---

## Field-Level Security

Restrict:

- salary
- financial data
- private records

---

# 4.6 WORKFLOW ENGINE

## Purpose

Manage business processes.

---

# Features

## State Machines

Example:

```text id="3g7mp4"
Draft → Confirmed → Approved → Done
```

---

## Approval Chains

Support:

- manager approval
- multi-level approval

---

## Notifications

Trigger:

- email
- websocket
- alerts

---

# 4.7 AI AGENT ENGINE

The ERP must become AI-native.

---

# AI Agent Roles

## Assistant Agent

Helps users:

- navigate ERP
- answer questions
- generate reports

---

## Automation Agent

Executes:

- repetitive tasks
- invoice creation
- reminders

---

## Analytics Agent

Provides:

- predictions
- insights
- anomalies

---

## Workflow Agent

Participates in:

- approvals
- escalations

---

# AI Safety Rules

AI:

- cannot bypass ACL
- must log actions
- must require approval for critical operations

---

# Natural Language ERP

Example:

```text id="v7s0a6"
"Create invoice for customer ABC"
```

AI converts:

- intent
- model action
- ORM operation

---

# 4.8 WEB CLIENT ENGINE

Frontend without React/Vue.

---

# Features

## Dynamic Routing

Support:

- menus
- navigation
- breadcrumbs

---

## RPC Layer

Frontend communicates using JSON-RPC.

---

## Dynamic Forms

Forms update asynchronously.

---

## Realtime Updates

Support:

- websocket
- notifications
- live dashboards

---

# 4.9 REPORTING ENGINE

## Features

- PDF reports
- Excel exports
- dashboard widgets
- analytics

---

# 4.10 DEVELOPER PLATFORM

## Goal

Enable ecosystem growth.

---

# CLI Tools

```bash id="o7wz2w"
sumeru addon create sales
```

---

# SDK

Developers should easily:

- create models
- define views
- register actions

---

# Testing Framework

Support:

- unit tests
- module tests
- integration tests

---

# 5. DEVELOPMENT ROADMAP

---

# PHASE 1 — FOUNDATION

## Build

- HTTP server
- PostgreSQL
- XML parser
- HTML renderer

---

# PHASE 2 — BASIC ORM

## Build

- model registry
- CRUD
- migrations

---

# PHASE 3 — MODULE ENGINE

## Build

- addon loader
- manifests
- dependency graph

---

# PHASE 4 — FORM VIEW ENGINE

## Build

- XML forms
- renderer
- form submission

---

# PHASE 5 — ACTION SYSTEM

## Build

- buttons
- dispatchers
- workflows

---

# PHASE 6 — TREE + KANBAN

## Build

- lists
- cards
- filtering

---

# PHASE 7 — SECURITY

## Build

- users
- ACL
- record rules

---

# PHASE 8 — WEB CLIENT

## Build

- JS runtime
- SPA navigation
- RPC

---

# PHASE 9 — AI AGENTS

## Build

- assistant APIs
- automation
- analytics

---

# PHASE 10 — MARKETPLACE

## Build

- addon registry
- installer
- version manager

---

# 6. LONG-TERM VISION

Eventually the platform should support:

- CRM
- Sales
- Inventory
- Accounting
- Manufacturing
- HR
- Payroll
- POS
- eCommerce
- Mobile apps
- AI workflows
- SaaS multi-tenancy

---

# 7. FINAL ARCHITECTURE GOAL

```text id="5r0bq3"
Go Backend
+
XML Declarative UI
+
Plug-and-Play Addons
+
AI-Native ERP
+
Scalable Enterprise Framework
```

---

# 8. GUIDING PRINCIPLES

1. Everything modular
2. Addons plug-and-play
3. UI declarative
4. Security centralized
5. AI permission-aware
6. Backend strongly typed
7. Workflows configurable
8. Extensibility over shortcuts
9. Community ecosystem first
10. Developer experience matters
