# Sumeru Evolution Plan: AI Integration & Structural Refactor
> Merged Roadmap for Core Cleanup, Naming Migration, and AI Capability

This definitive plan outlines the steps to simplify the `sumeru` core, adopt Go-centric naming conventions, and implement advanced AI features.

---

## 1. Unified Priority Order

| Priority | Phase | Goal |
| :--- | :--- | :--- |
| **1. CRITICAL** | **Core Cleanup & Naming** | [DONE] Strip `core/` of logic. Framework metadata uses **`sys.*`**; shared business entities use **`core.*`**. |
| **2. HIGH** | **Branding & Pure Go** | ✅ **Complete** | Standardize namespaces (sys.*, core.*) and fields (core_id, core_model). |
| **3. HIGH** | **Base Consolidation** | [DONE] Create a unified `addons/base` module using the new naming schema. |
| **4. MEDIUM** | **Framework Hooks** | [DONE] Implement UI and ORM extensibility for AI and 3rd party plugins. |
| **5. MEDIUM** | **sumeru_ai Suite** | [IN PROGRESS] Build the AI Assistant, NL2D Search, and Smart Summaries. |

---

## 2. Phase 1: Core Framework Cleanup & Naming Migration

The framework (`core/`) should only contain the engine. We will also adopt Go-centric naming to improve clarity.

### Structural Refactor
- **[MOVE]** `core/base/user` -> `addons/base_user`
- **[MOVE]** `core/base/company` -> `addons/base_company`
- **[MOVE]** `core/base/settings` -> `addons/base_settings`
- **[RESULT]** `sumeru/core` becomes a pure, lightweight ERP platform.

### Naming Migration
Adopt a clean, explicit namespace for all models:
- **`sys.*`** (System): Framework metadata (views, rules, models).
- **`core.*`** (Core): Shared business entities (users, partners, companies).

---

## 3. Phase 3: Unified `base` Addon

Create `addons/base` as the definitive foundation for all Sumeru installations.

- **Foundational Models**: `core.partner`, `core.user`, `core.company`, `core.group`.
- **System Metadata**: Formally implement `sys.model` and `sys.field` models. This allows AI to "discover" the schema.
- **Security**: Move bootstrap security logic into `addons/base/security/` using the new `sys.` names.

---

## 4. Phase 4: Framework Extensibility (Hooks)

Enable the core to be extended without modification.

- **UI Hooks**: Registries in `core/engine/render/render_types.go` (`ShellHooks`, `NotebookHooks`) with **`RegisterShellHook`** / **`RegisterNotebookHook`** for injecting shell and notebook HTML without editing core templates.
- **ORM Interceptors**: Hooks in `orm.Search` to allow the AI module to translate natural language into ORM domains.

---

## 5. Phase 5: Sumeru AI Implementation

Build AI capabilities on top of the clean foundation.

- **NL2D Search**: "Show me orders from last week" -> Converts to ORM Domain.
- **Smart Chatter**: AI-powered summaries of message threads.
- **AI Assistant**: A persistent chat widget for record insights and navigation help.

---

## 6. Documentation Strategy

Maintain separate documentation tracks under `sumeru/docs/`:

### Developer Documentation (`/docs/developer/`)
- **Naming Conventions**: Detailed guide on `sys.` vs `core.` and addon structure.
- **ORM API**: Documentation on CRUD, Domains, and Interceptors.
- **UI Architecture**: Guide on XML views and UI Hooks.

### User Documentation (`/docs/user/`)
- **Core Concepts**: Understanding Users, Companies, and Partners.
- **AI Features**: How to use NL2D search and the AI Assistant.
- **Navigation**: Shell and Workspace walkthrough.
