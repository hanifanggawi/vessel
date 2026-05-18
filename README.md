# Vessel

*Seal the blinding light that plagues their dreams*

---

Vessel is a cross-platform CLI for blocking and limiting website usage, it works by managing rules set on a config file and reflected on the user's `hosts` file. The rule blocking is enforced by a background daemon that constantly reads and syncs the hosts file. Seal the vessel to make it harder to remove rules. 

Supported on Linux and Windows.

## Installation

### Linux (install script)

```sh
curl -fsSL https://raw.githubusercontent.com/hanifanggawi/vessel/main/install.sh | sh
```

This downloads the latest release, verifies its checksum, and installs the
binary to `/usr/local/bin`. Then:

```sh
vessel init                  # create your per-user config
sudo vessel daemon install   # run the reconcile daemon at startup (systemd)
```

### Windows

Download the latest `vessel_*_windows_*.zip` from the
[releases page](https://github.com/hanifanggawi/vessel/releases), extract
`vessel.exe`, and place it on your `PATH`. Then, in an **Administrator**
prompt:

```powershell
vessel init
vessel daemon install        # registers a startup task (runs as SYSTEM)
```

### From source

```sh
go install github.com/hanifanggawi/vessel@latest
```

> Editing the `hosts` file requires elevation: run write commands with `sudo`
> on Linux or from an Administrator prompt on Windows. Vessel resolves the
> same per-user config whether or not it is invoked via `sudo`.

## Usage

```sh
# Block a domain permanently
sudo vessel add youtube.com

# Block only during a recurring time window (repeatable, merges in)
sudo vessel add youtube.com --window 09:00-17:00

# Block for a duration from now
sudo vessel add youtube.com --for 2h30m

# Block until a time of day today
sudo vessel add youtube.com --until 18:30

# Overwrite an existing rule for a domain
sudo vessel add youtube.com --for 1h --replace

# List configured rules
vessel list

# Remove a rule (asks for a challenge if the vessel is sealed)
sudo vessel release instagram.com
```

### Sealing

```sh
vessel seal       # make it harder to release a rule by requiring a challenge
vessel unseal     # requires passing the challenge while sealed
```

### Daemon

```sh
sudo vessel daemon install     # enable at startup (systemd / scheduled task)
sudo vessel daemon uninstall   # remove the startup service

vessel daemon start            # ad-hoc background run (not boot-managed)
vessel daemon status
vessel daemon stop
vessel daemon restart
```

When installed as a service, manage it with `systemctl status vessel`
(Linux) or `schtasks /Query /TN vessel` (Windows).

## Configuration

| Variable | Purpose |
| --- | --- |
| `VESSEL_DOMAINS_CONFIG_PATH` | Override the rules config location. Defaults to `~/.config/vessel/domainconfig.toml` (the invoking user's, even under `sudo`). |
| `VESSEL_HOSTS_PATH` | Override the hosts file path (defaults to the OS standard). |
