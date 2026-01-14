# Installation

This document describes how to install germinator on your system.

## Quick Install

The easiest way to install germinator is using our install script:

```bash
curl -sSL https://gitlab.com/amoconst/germinator/-/raw/main/install.sh | bash
```

This will automatically detect your operating system and architecture, fetch the latest version from GitLab releases, and install the appropriate binary.

## Manual Installation

For manual installation, download a specific version from the [GitLab releases page](https://gitlab.com/amoconst/germinator/-/releases). Replace `0.3.0` with your desired version.

### Linux

#### Linux (AMD64)

```bash
curl -L https://gitlab.com/amoconst/germinator/-/releases/0.3.0/downloads/germinator_0.3.0_linux_amd64.tar.gz -o germinator.tar.gz
tar -xzf germinator.tar.gz
sudo mv germinator /usr/local/bin/
```

#### Linux (ARM64)

```bash
curl -L https://gitlab.com/amoconst/germinator/-/releases/0.3.0/downloads/germinator_0.3.0_linux_arm64.tar.gz -o germinator.tar.gz
tar -xzf germinator.tar.gz
sudo mv germinator /usr/local/bin/
```

### macOS

#### macOS (Intel)

```bash
curl -L https://gitlab.com/amoconst/germinator/-/releases/0.3.0/downloads/germinator_0.3.0_darwin_amd64.tar.gz -o germinator.tar.gz
tar -xzf germinator.tar.gz
sudo mv germinator /usr/local/bin/
```

#### macOS (Apple Silicon)

```bash
curl -L https://gitlab.com/amoconst/germinator/-/releases/0.3.0/downloads/germinator_0.3.0_darwin_arm64.tar.gz -o germinator.tar.gz
tar -xzf germinator.tar.gz
sudo mv germinator /usr/local/bin/
```

### Windows

#### Windows (AMD64)

```powershell
# Using PowerShell
Invoke-WebRequest -Uri "https://gitlab.com/amoconst/germinator/-/releases/0.3.0/downloads/germinator_0.3.0_windows_amd64.zip" -OutFile "germinator.zip"
Expand-Archive -Path germinator.zip -DestinationPath .
Move-Item germinator.exe C:\Windows\System32\
```

Or download the `.zip` file manually from the [releases page](https://gitlab.com/amoconst/germinator/-/releases), extract it, and add the `germinator.exe` to your PATH.

## Checksum Verification

To verify the integrity of downloaded files, you can check the SHA256 checksum:

### Linux/macOS

```bash
# Download the checksums file for version 0.3.0
curl -L https://gitlab.com/amoconst/germinator/-/releases/0.3.0/downloads/checksums.txt -o checksums.txt

# Verify the checksum
sha256sum -c checksums.txt
```

### Windows

```powershell
# Download the checksums file
Invoke-WebRequest -Uri "https://gitlab.com/amoconst/germinator/-/releases/latest/downloads/checksums.txt" -OutFile "checksums.txt"

# Verify the checksum (requires certutil)
certutil -hashfile germinator.tar.gz SHA256
# Compare the output with the corresponding line in checksums.txt
```

## GPG Signature Verification (Optional)

To verify the authenticity of the release using GPG signatures:

```bash
# Download the signature file
curl -L https://gitlab.com/amoconst/germinator/-/releases/latest/downloads/checksums.txt.sig -o checksums.txt.sig

# Import the signing key (if you haven't already)
# Note: You'll need to import the public key from the project maintainer
gpg --keyserver keyserver.ubuntu.com --recv-keys SIGNING_KEY_FINGERPRINT

# Verify the signature
gpg --verify checksums.txt.sig checksums.txt
```

## Verify Installation

After installation, verify that germinator is working correctly:

```bash
germinator version
```

You should see output like:

```
germinator 0.2.0 (abc1234) 2025-01-13
```

## SBOM (Software Bill of Materials)

Each release includes an SBOM (Software Bill of Materials) in SPDX format. You can download it from the [releases page](https://gitlab.com/amoconst/germinator/-/releases):

```
germinator_0.2.0_sbom.spdx.json
```

The SBOM provides a complete list of dependencies used in the build, which is useful for security and compliance purposes.

## Uninstall

### Linux/macOS

```bash
sudo rm /usr/local/bin/germinator
```

### Windows

```powershell
Remove-Item C:\Windows\System32\germinator.exe
```

## Next Steps

After installation, check out the [README](README.md) for usage instructions, or run:

```bash
germinator --help
```
