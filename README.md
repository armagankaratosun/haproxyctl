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

`haproxyctl` aims to provide a CLI for managing HAProxy resources using the [HAProxy Data Plane API](https://www.haproxy.com/documentation/dataplaneapi/community/).

### Current Commands

| Category        | Command Example                                     | Description |
|-----------------|-----------------------------------------------------|---|
| Configuration   | `haproxyctl get configuration`                     | Fetch full HAProxy configuration |
| Configuration   | `haproxyctl get configuration version`              | Fetch configuration version |
| Backends        | `haproxyctl get backends`                           | List all backends |
| Backends        | `haproxyctl get backends <name>`                    | Show a specific backend |
| Backends        | `haproxyctl describe backends <name>`               | Show backend details + servers |
| Backends        | `haproxyctl create backend <name>`                  | Create backend (from flags or YAML file) |
| Backends        | `haproxyctl delete backends <name>`                  | Delete a backend |
| Servers         | `haproxyctl get servers <backend>`                  | List servers in a backend |
| Servers         | `haproxyctl create server <backend>`                | Add server to backend |
| Servers         | `haproxyctl delete server <backend> <server>`       | Remove server from backend |
| Frontends       | `haproxyctl get frontends`                          | List all frontends |
| Frontends       | `haproxyctl get frontends <name>`                   | Show frontend details |
| ACLs            | `haproxyctl get acls <frontend>`                    | List ACLs for a frontend |
| Service         | `haproxyctl service reload`                         | Reload HAProxy configuration |

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

