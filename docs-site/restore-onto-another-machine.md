# Restore onto another machine

Your laptop is gone. The Restic repository is still there. This page walks through getting
the files back on a different machine, using only the commands that exist in the current
build.

The trick is the repository mount point: a directory whose `.snapshot` entry shows the whole
repository, so you can browse and copy from a backup that was taken on a host you no longer
have.

## 1. Install Snapback on the new machine

You need FUSE (fuse3 on Linux, macFUSE on macOS), the `restic` command on your `PATH`, and
the Restic repository plus its password.

```
curl -fsSL https://snapback.run/install.sh | sh
```

## 2. Run setup and pick the mount point

Point setup at the repository and give it a mount point:

```
snapback setup --repo /srv/restic/home-nas \
  --password-file ~/.config/restic/home-nas.pw \
  --mount /mnt/home-nas
```

Leave `--mount` off and interactive setup asks one question for the mount point, with the
default already filled in — press Enter to accept it.

The default mount point is `/mnt/<repository id>` on Linux; on macOS it is the per-user path
`~/Library/Application Support/snapback/mounts/<repository id>`.
Under config schema v1 the instance name is the repository id, so the directory is named
after the repository itself. Named instances are a schema v2 idea and are not implemented
yet, so one repository gets one mount point.

Setup records the choice as `repositories[0].mount_point` in the config file:

```yaml
repositories:
  - id: home-nas
    repository: /srv/restic/home-nas
    password_file: /home/you/.config/restic/home-nas.pw
    mount_point: /mnt/home-nas
```

`--mount ""` writes an empty value and leaves the repository unmounted.

## 3. Start the daemon

```
snapback run
```

Once the repository is ready, the daemon creates the managed `.snapshot` link inside the
mount point directory.

## 4. Browse the repository

```
ls /mnt/home-nas/.snapshot/ids/
```

Each entry under `.snapshot/ids/` is one snapshot, keyed by its Restic id, holding the paths
that snapshot captured. Snapback mounts the backend with `restic mount` under
`--path-template ids/%I`, so `ids/` is the only view it publishes.

This view is the whole repository, whatever machine took the snapshot. It deliberately
ignores your `roots` and `prefix_map` config, which only shape the per-directory `.snapshot`
links on a machine that is still being backed up.

## 5. Copy the files out

The mount is read-only: restoring never writes to the Restic repository, and neither can you
through this path. A restore is therefore a plain copy out of it:

```
cp -a /mnt/home-nas/.snapshot/ids/4f2a8b1c/home/alice/project ~/project
```

Use `cp -a` (or `rsync -a`) so permissions, timestamps and symlinks survive the copy.

## 6. Check the state when something is off

```
snapback status
```

The mount point line reports one of four states:

| State | Meaning |
|---|---|
| `linked` | the managed `.snapshot` link is in place and resolves — you can browse |
| `missing` | the link is absent or dangling; run `snapback run` |
| `disabled` | this repository has no `mount_point` set |
| `conflict` | something that is not a managed link already sits at that path; move it aside |

If the daemon logs a `mount point link failed` warning, the mount point directory usually
does not exist yet. Create it and let the daemon retry:

```
sudo mkdir -p /mnt/home-nas
```

## 7. Stop when you are done

Stopping the daemon removes the `.snapshot` link and keeps the mount point directory itself,
so nothing is left dangling and the next `snapback run` picks up where it left off.
