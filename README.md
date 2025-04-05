# LinkWinBT

[![License](https://img.shields.io/github/license/vvoland/linkwinbt)](LICENSE)


LinkWinBT is a utility that extracts the Bluetooth pairing from the Windows registry and applies it to your Linux Bluetooth configuration.

It allows to have the same Bluetooth device paired on both Linux and Windows.



https://github.com/user-attachments/assets/505bd515-0ed4-4e51-adb1-58dc663526bd



## Problem

When using a dual-boot system with Linux and Windows, Bluetooth devices cannot be used seamlessly across both operating systems. This is because each OS generates and stores a unique link key for the same Bluetooth device but use the same Bluetooth controller so the linked device can't distinguish between the two systems.

## Solution

LinkWinBT extracts the Bluetooth link key from the Windows registry and applies it to your Linux Bluetooth configuration, allowing your Bluetooth devices to work seamlessly across both operating systems.

You can read more about the exact steps in the [DESIGN.md](DESIGN.md) file.

## Usage


### Prerequisites
- A dual-boot Linux/Windows system
- Bluetooth devices paired on **Linux FIRST** and then on Windows
- Access to the Windows registry file (`C:\Windows\System32\config\SYSTEM`) 
- [non-Docker usage] `reged` tool installed on Linux (usualy provided by `chntpw` package)


### Easy run with Docker

```bash
# Pair your bluetooth device on Linux first
# Reboot to Windows and pair it there
# Reboot to Linux

# Mount your Windows partition
$ sudo mount /dev/sdXY /mnt

# Run the container
$ docker run --rm -it -v /mnt:/windows:ro -v /var/lib/bluetooth:/var/lib/bluetooth vlnd/linkwinbt /windows

# Restart bluetooth
$ sudo systemctl restart bluetooth
```

> **Note:** You can further restrict the container to limit its access to your system.

```bash
# No network access, mount only one registry file
$ docker run --rm -it \
  --network none \   # No network access
  -v /mnt/Windows/System32/config/SYSTEM:/windows:ro \
  -v /var/lib/bluetooth:/var/lib/bluetooth \
  vlnd/linkwinbt /windows


# Same as above but only print the link key
# You will need to modify the bluetooth files manually
$ docker run --rm -it \
  --network none \
  -v /mnt/Windows/System32/config/SYSTEM:/windows:ro \
  -v /var/lib/bluetooth:/var/lib/bluetooth:ro \
  vlnd/linkwinbt -dry /windows
```


### Manual

1. Install

```bash
go install grono.dev/linkwinbt/cmd/linkwinbt@latest
```

1. Run LinkWinBT with the path to your Windows installation or SYSTEM registry file:

```bash
sudo linkwinbt /path/to/windows
```

or

```bash
sudo linkwinbt /path/to/windows/System32/config/SYSTEM
```

3. If multiple Bluetooth controllers or devices are found, select the appropriate one from the list
4. The tool will extract the Windows link key and apply it to your Linux configuration
5. Bluetooth service will be restarted automatically
