# Welcome to Sumeru ERP

Sumeru is a modern, modular ERP system designed for simplicity and speed. This guide will help you understand the basic concepts of how the system is organized.

## Core Concepts

### The Workspace
The primary area where you interact with data. Whether you are looking at a list of customers or editing an invoice, you are working in the **Workspace** (URLs under **`/web`**, e.g. list and form views).

### Home, Apps, and Settings
- **Home** (`/web/home`): quick access to installed **application** modules (app hub).
- **Apps** (`/web/apps`): full catalog, filters, install/update actions, and module detail.
- **Settings** (`/web/settings`): overview of configuration menus from the Settings tree plus shortcuts to installed apps; deeper items open in the normal workspace.

### Applications (Apps)
Sumeru is modular. Different business functions like **Sales**, **CRM**, and **Inventory** are separate Apps. You reach them from **Home** or the top bar after choosing **Apps** or a pinned module.

### Partners, Users, and Companies
The system centers around three main types of entities:
- **Companies**: The legal organizations that own the data in the system.
- **Users**: The people who log into Sumeru to do work.
- **Partners**: Everyone your business interacts with—this includes customers, vendors, and even your own employees.

## Natural Language Search
Sumeru features advanced AI-driven search. Instead of clicking through complex filters, you can often just type what you are looking for in plain English, like *"Show me my open leads from last week"*.

## The Activity Panel
On the right side of most records, you will see an **Activity Panel**. This is where you can see the history of changes, send messages to team-mates, and see AI-generated summaries of long conversations.

---

**Developers:** see **[`docs/developer/index.md`](../developer/index.md)** and **[`docs/howtos/index.md`](../howtos/index.md)** for extending Sumeru (addons, views, security).
