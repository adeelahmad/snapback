# Docker

Snapback ships a container image alongside the release tarballs.
The image bundles the `snapback` binary, restic 0.18.0 and fuse3, so nothing extra is needed
on the host: the container only needs `/dev/fuse` and the directories you want a history entry
in.

## The image

| What | Value |
|---|---|
| Name | `ghcr.io/adeelahmad/snapback` |
| Tags | `<version>` for each release, plus `latest` |
| Architectures | `linux/amd64` and `linux/arm64` |
| Base | `alpine:3.21`, pinned by digest |
| Entrypoint | `/snapback`, so container arguments are Snapback subcommands |

Everything inside is pinned: the base image by digest, `fuse3=3.16.2-r1` by apk version, and
the restic 0.18.0 download by the SHA256 from restic's own `SHA256SUMS`. The `snapback` binary
is not built in the image; goreleaser hands the release binary to the build, so the image and
the tarball for the same architecture carry the exact same bytes.

Images are published by the release workflow, which builds both architectures with buildx and
pushes them on a tag. No image has been pushed yet: the first tagged release after this page
is the first one with an image at `ghcr.io/adeelahmad/snapback`.

## Running it

Snapback mounts a FUSE file system, so the container needs three things a default `docker run`
does not give it: `--device /dev/fuse` for the FUSE device, `--cap-add SYS_ADMIN` for the mount
syscall, and `--security-opt apparmor:unconfined` on hosts where AppArmor blocks FUSE mounts.
The directories it watches are bind-mounted with `:rshared` so the mounts Snapback makes inside
the container propagate back to the host.

```bash
docker run -d --name snapback \
  --device /dev/fuse \
  --cap-add SYS_ADMIN \
  --security-opt apparmor:unconfined \
  -v /home/alex/.config/snapback:/config \
  -v /home/alex/.local/state/snapback:/state \
  -v /home/alex/work:/home/alex/work:rshared \
  ghcr.io/adeelahmad/snapback:latest \
  run --config /config/config.yaml
```

The `.snapshot` links appear inside the bind-mounted host directories, so you browse history
with the same `cd` and `ls` you would use without a container.

A few notes on the mounts above:

- The config and state directories are bind-mounted so the configuration, the catalog and the
  managed links survive a container restart. Point `state_dir` at the container path.
- Each directory under `roots` needs its own `:rshared` bind mount, at the same path inside the
  container as on the host, so the paths Snapback records match what you see outside.
- A bind mount made without `:rshared` still gets its `.snapshot` link, but the FUSE mount
  behind it stays invisible outside the container, so the link reads as empty on the host.

## Repository mount points

`repositories[i].mount_point` behaves the same way inside a container as outside: Snapback
mounts the repository there, and the directory has to be a `:rshared` bind mount for the
result to be visible on the host. Give the mount point its own `-v /mnt/home-nas:/mnt/home-nas:rshared`
and leave the key set to that path.

## Credentials

The container reads the same configuration as a host install, so the repository password file
and any `environment:` values are read from inside the container. Bind-mount the password file
read-only and give its container path to `password_file`; see
[Configuration](configuration.md) for the full key reference.
