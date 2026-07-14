# NoCloud cloud-init seed for ci-pilot

`user-data` + `meta-data` are packed into a CIDATA-labelled ISO
(`seed.iso`) and attached as a second optical drive. The Ubuntu cloud
image's cloud-init consumes them on first boot: creates the `runner`
account (locked password, key-only, passwordless sudo), installs sshd,
and injects the committed SSH **public** key.

Rebuild the ISO after editing (host has no genisoimage; a small Go
`iso9660` writer is used):

```bash
mkseed <seed-dir> <out.iso>   # writes with volume label CIDATA
```

The `ssh_authorized_keys` entry is a public key — safe to commit. The
matching private key lives only in the host operator's `~/.ssh`.
