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

`haproxyctl` aims to provide a CLI for managing HAProxy resources using the [HAProxy Data Plane API v3](https://www.haproxy.com/documentation/dataplaneapi/community/).

The UX is intentionally “kubectl‑ish”: you can `get`, `create`, `edit`, and `delete` resources, and work with YAML manifests that include `apiVersion` and `kind`.

### Current Commands

| Category        | Command Example                                          | Description |
|-----------------|----------------------------------------------------------|---|
| Auth            | `haproxyctl login`                                       | Configure Data Plane API URL and credentials (~/.config/haproxyctl) |
| Configuration   | `haproxyctl get configuration version -o json`           | Fetch configuration version |
| Configuration   | `haproxyctl get configuration raw`                       | Fetch raw HAProxy configuration |
| Backends        | `haproxyctl get backends`                                | List all backends (sorted by name) |
| Backends        | `haproxyctl get backends <name>`                         | Show a specific backend (includes servers column) |
| Backends        | `haproxyctl describe backends <name>`                    | Show backend details + servers (descriptive view) |
| Backends        | `haproxyctl create backends <name> [flags]`              | Create backend (from flags) |
| Backends        | `haproxyctl create -f examples/backend-with-server.yaml` | Create backend + servers from a YAML manifest (`kind: Backend`) |
| Backends        | `haproxyctl edit backends <name>`                        | Edit backend + its servers in `$EDITOR` via manifest |
| Backends        | `haproxyctl delete backends <name>`                      | Delete a backend |
| Servers         | `haproxyctl get servers <backend>`                       | List servers in a backend (sorted by name) |
| Servers         | `haproxyctl get servers <backend> <server> -o yaml`      | Show a specific server as a manifest (`kind: Server`) |
| Servers         | `haproxyctl create servers <backend> <server> [...]`     | Add server to backend (flags) |
| Servers         | `haproxyctl create -f examples/server.yaml`              | Create a server from a YAML manifest |
| Servers         | `haproxyctl delete server <backend> <server>`            | Remove server from backend |
| Frontends       | `haproxyctl get frontends`                               | List all frontends (sorted by name) |
| Frontends       | `haproxyctl get frontends <name>`                        | Show frontend details (includes binds column) |
| Frontends       | `haproxyctl create frontends <name> [...]`               | Create a frontend and optional binds (flags) |
| Frontends       | `haproxyctl create -f examples/frontend-with-binds.yaml` | Create a frontend + binds from a YAML manifest |
| Frontends       | `haproxyctl edit frontends <name>`                       | Edit frontend + its binds in `$EDITOR` via manifest |
| Frontends       | `haproxyctl delete frontends <name>`                     | Delete a frontend |
| ACLs            | `haproxyctl get acls <frontend>`                         | List ACLs for a frontend |

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

## ⚠️ Important Notice: Not the Same as Other `haproxyctl` Tools

This project **`haproxyctl`** is a **new, independent implementation** designed specifically to interact with the [HAProxy Data Plane API](https://www.haproxy.com/documentation/dataplaneapi/community/).  

It has **no relation** to the other `haproxyctl` tools like:

- [`haproxyctl` package from Ubuntu](https://ubuntuupdates.org/package/core/jammy/universe/base/haproxyctl)
- [`haproxyctl` from easybiblabs](https://github.com/easybiblabs/haproxyctl)

These tools interact directly with HAProxy **through Unix sockets**, while **this version communicates purely via HTTP REST calls to the HAProxy Data Plane API**. 

If you’re looking for socket-based HAProxy control, those other tools might be what you want.  
If you need a **Data Plane API management CLI** for HAProxy, this project is the right tool for the job.

---

## License

Apache License 2.0

## Authors

- **Armagan Karatosun** — *Supreme Overlord of haproxyctl, First of His Name.*

## Special Thanks
-  **ChatGPT** — *AI Sidekick, Tireless Rubber Duck, Suggestor of Both Useful and Questionable Code, and Writer of This Very Sentence.*
