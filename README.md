# haproxyctl
A Simple command line utility to manage HAProxy instances with HAProxy Data Plane API

⚠️ **Warning - Here Be Dragons!**

This project is **fresh out of the oven** and still **half-baked**.

`haproxyctl` is under **heavy construction**, so things might break, vanish, or completely change **while you're looking at them**.

**Always** back up your HAProxy configuration before you let `haproxyctl` poke around — or risk being chased by angry sysadmins.  

Also, I’m still figuring out how I want to shape these APIs, so they **could mutate at any moment**.  

**TL;DR:** This is **NOT production-ready**. If you use it in production and your HAProxy starts speaking in tongues, **that’s on you**.

---

## Features

`haproxyctl` aims to provide a CLI for managing HAProxy resources using the [HAProxy Data Plane API v3 (community edition)](https://www.haproxy.com/documentation/dataplaneapi/community/).

The UX is intentionally “kubectl‑ish”: you can `get`, `create`, `edit`, and `delete` resources, and work with YAML manifests that include `apiVersion` and `kind`.

### Scope & Compatibility

- Targets **HAProxy Data Plane API v3** (community). The CLI normalizes your configured URL to include `/v3` if you omit it.
- Focuses on configuration workflows that map cleanly to the Data Plane API:
  - Backends, frontends, servers (list, describe, create, edit, delete, apply).
  - ACLs (read‑only listing per frontend).
  - Selected configuration sections (`globals`, `defaults`) with `get`, `edit`, and manifest‑driven `apply`.
- Does **not** (yet) cover:
  - Certificate management, stick tables, service discovery configuration.
  - Enterprise‑only features (e.g. integrated Git in the Data Plane API).
  - Socket‑level HAProxy control (this CLI always talks HTTP to Data Plane API).

### Current Commands

| Category        | Command Example                                          | Description |
|-----------------|----------------------------------------------------------|---|
| Auth            | `haproxyctl login`                                       | Configure Data Plane API URL and credentials (~/.config/haproxyctl) |
| Configuration   | `haproxyctl get configuration version -o json`           | Fetch configuration version (JSON by default) |
| Configuration   | `haproxyctl get configuration raw`                       | Fetch raw HAProxy configuration |
| Configuration   | `haproxyctl get configuration globals`                   | Show structured view of global settings (when available) |
| Configuration   | `haproxyctl edit configuration globals`                  | Open a manifest‑style `Global` section in `$EDITOR` |
| Configuration   | `haproxyctl get configuration defaults <name>`           | Show a specific `Defaults` section (table / YAML / JSON) |
| Configuration   | `haproxyctl edit configuration defaults <name>`          | Edit a named `Defaults` section in `$EDITOR` |
| Backends        | `haproxyctl get backends`                                | List all backends (sorted by name) |
| Backends        | `haproxyctl get backends <name>`                         | Show a specific backend (includes servers column) |
| Backends        | `haproxyctl describe backends <name>`                    | Show backend details + servers (descriptive view) |
| Backends        | `haproxyctl create backends <name> [flags]`              | Create backend (from flags) |
| Backends        | `haproxyctl create -f examples/backend-with-server.yaml` | Create backend + servers from a YAML manifest (`kind: Backend`) |
| Backends        | `haproxyctl edit backends <name>`                        | Edit backend + its servers in `$EDITOR` via manifest |
| Backends        | `haproxyctl delete backends <name>`                      | Delete a backend |
| Backends        | `haproxyctl apply -f backend.yaml`                       | Create or replace a backend from a manifest |
| Servers         | `haproxyctl get servers <backend>`                       | List servers in a backend (sorted by name) |
| Servers         | `haproxyctl get servers <backend> <server> -o yaml`      | Show a specific server as a manifest (`kind: Server`) |
| Servers         | `haproxyctl create servers <backend> <server> [...]`     | Add server to backend (flags) |
| Servers         | `haproxyctl create -f examples/server.yaml`              | Create a server from a YAML manifest |
| Servers         | `haproxyctl delete server <backend> <server>`            | Remove server from backend |
| Servers         | `haproxyctl apply -f server.yaml`                        | Create or replace a server from a manifest |
| Frontends       | `haproxyctl get frontends`                               | List all frontends (sorted by name) |
| Frontends       | `haproxyctl get frontends <name>`                        | Show frontend details (includes binds column) |
| Frontends       | `haproxyctl create frontends <name> [...]`               | Create a frontend and optional binds (flags) |
| Frontends       | `haproxyctl create -f examples/frontend-with-binds.yaml` | Create a frontend + binds from a YAML manifest |
| Frontends       | `haproxyctl edit frontends <name>`                       | Edit frontend + its binds in `$EDITOR` via manifest |
| Frontends       | `haproxyctl delete frontends <name>`                     | Delete a frontend |
| Frontends       | `haproxyctl apply -f frontend.yaml`                      | Create or replace a frontend from a manifest |
| ACLs            | `haproxyctl get acls <frontend>`                         | List ACLs for a frontend |
| Apply           | `haproxyctl apply -f <file>`                             | Generic manifest‑driven apply (Backend, Frontend, Server, Global, Defaults) |

---

## Installation

### Build from Source

```sh
git clone https://github.com/armagankaratosun/haproxyctl.git
cd haproxyctl
go build -o haproxyctl
./haproxyctl --help
```

### Prebuilt Binary (future)

Once stable releases are available, binaries will be published under [Releases](https://github.com/armagankaratosun/haproxyctl/releases).

Example installation for Linux:

```sh
curl -L -o haproxyctl https://github.com/armagankaratosun/haproxyctl/releases/download/v0.1.0/haproxyctl-linux-amd64
chmod +x haproxyctl
sudo mv haproxyctl /usr/local/bin/
```

## Getting Started

1. **Configure access to the Data Plane API (v3)**  

   ```sh
   haproxyctl login
   ```

   This will prompt for:

   - API base URL (e.g. `http://10.0.0.187:5555` or `http://10.0.0.187:5555/v3`)
   - Username
   - Password

   The URL is automatically normalized to include `/v3` if you omit a version, and credentials are stored in `~/.config/haproxyctl/config.json`.

2. **Explore resources**

   ```sh
   haproxyctl get backends
   haproxyctl get frontends
   haproxyctl get servers <backend>
   ```

3. **Output formats**

   By default, `haproxyctl get ...` prints **tables**:

   - Object lists (`get backends`, `get frontends`, `get servers <backend>`) are shown as tables with a stable, name‑sorted order.
   - Single objects (`get backends <name>`, `get servers <backend> <server>`) are shown as a one‑row table.

   You can request structured output explicitly:

   ```sh
   haproxyctl get backends -o yaml
   haproxyctl get servers mybackend -o yaml
   ```

   - `-o yaml` / `-o json` returns manifest‑style objects with `apiVersion` and `kind`.
   - For lists (e.g. `get servers mybackend -o yaml`), output uses a `kind: List` wrapper with `items: [...]`, similar to Kubernetes.
   - For configuration sections:
     - `get configuration globals` prints a table when the Data Plane API returns meaningful JSON.
     - If the v3 API returns an empty `{}` for globals in your setup, the CLI prints a hint:

       > configuration/globals no rules defined; use 'haproxyctl get configuration raw' and 'haproxyctl create configuration raw' for global settings

     - `get configuration defaults <name>` prints a table row by default; `-o yaml/json` returns a `Defaults` manifest including the `name`.

3. **Work with manifests (optional, kubectl‑style)**

   - Create from YAML:

     ```sh
     haproxyctl create -f examples/backend-with-server.yaml
     haproxyctl create -f examples/server.yaml
     ```

   - Edit live resources in your editor:

     ```sh
     haproxyctl edit backends <backend-name>
     haproxyctl edit frontends <frontend-name>
     ```

   These commands open a manifest with `apiVersion: haproxyctl/v1` and `kind: Backend` / `kind: Frontend` where you can edit core fields and nested servers/binds.

4. **GitOps‑style workflow (manifests + git)**

   A common pattern is to treat manifests as the source of truth and let `haproxyctl` apply them:

   ```sh
   # Apply a backend + its servers from a manifest
   haproxyctl apply -f examples/backend-with-server.yaml

   # Tweak it live via edit
   haproxyctl edit backends example-backend-with-servers

   # Inspect local changes
   git diff
   ```

   The same approach works for configuration sections:

   ```sh
   # Edit a named defaults section (e.g. "unnamed_defaults_1")
   haproxyctl edit configuration defaults unnamed_defaults_1

   # Re‑apply from a manifest if desired
   haproxyctl apply -f defaults.yaml
   ```

   - `create` will fail with a 409 if the object already exists.
   - `apply` is **create‑or‑replace** and will tell you whether the resource was created, configured, or unchanged, using kubectl‑like messages (e.g. `backend/mybackend created`).

### Configuration notes

- On Data Plane API v3, the `/services/haproxy/configuration/global` endpoint may return an empty JSON object `{}` in some setups. In that case:
  - `haproxyctl get configuration globals` prints a friendly hint pointing you at `get configuration raw` / `create configuration raw`.
  - The raw configuration is the real source of truth for global options.
- `haproxyctl get configuration defaults <name>` and `edit configuration defaults <name>` operate on a named defaults section (e.g. `unnamed_defaults_1`). Defaults are not the same as a backend’s `default_backend`; they are their own configuration section.

## ⚠️ Important Notice: Not the Same as Other `haproxyctl` Tools

This project **`haproxyctl`** is a **new, independent implementation** designed specifically to interact with the [HAProxy Data Plane API](https://www.haproxy.com/documentation/dataplaneapi/community/).  

It has **no relation** to the other `haproxyctl` tools like:

- [`haproxyctl` package from Ubuntu](https://ubuntuupdates.org/package/core/jammy/universe/base/haproxyctl)
- [`haproxyctl` from easybiblabs](https://github.com/easybiblabs/haproxyctl)

These tools interact directly with HAProxy **through Unix sockets**, while **this version communicates purely via HTTP REST calls to the HAProxy Data Plane API**. 

If you’re looking for socket-based HAProxy control, those other tools might be what you want.  
If you need a **Data Plane API management CLI** for HAProxy, this project is the right tool for the job.

---

## Testing & Linting

This project uses standard Go tooling plus `golangci-lint` for static analysis.

- Run all unit tests:

  ```sh
  go test ./...
  ```

- Run tests for just the shared helpers:

  ```sh
  go test ./internal
  ```

- Optional compatibility smoke test:

  - `internal.TestCompatLiveDataPlaneAPI` performs a very lightweight check against a live Data Plane API.
  - It only runs if `haproxyctl` can load a local config (e.g. after `haproxyctl login`) and will `t.Skip` otherwise, so it is safe for environments without HAProxy.

- Run linters (if you have `golangci-lint` installed):

  ```sh
  golangci-lint run
  ```

  The configuration prefers practical checks; a few rules (e.g. some magic number and security warnings around CLI file paths) are intentionally relaxed where they would fight the UX.

## License

Apache License 2.0

## Authors

- **Armagan Karatosun** — *Supreme Overlord of haproxyctl, First of His Name.*

## Special Thanks
-  **ChatGPT** — *AI Sidekick, Tireless Rubber Duck, Suggestor of Both Useful and Questionable Code, and Writer of This Very Sentence.*
