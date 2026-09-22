# Launch checklist

The goal for launch day is a trending project of the day on GitHub. Trending is driven by
stars inside a 24–72 hour window, so everything a visitor sees in the first ten seconds —
the banner, the About line, the `cp` example — has to be right *before* the first post
goes out, not after.

Status is one of three values, and nothing else:

- `done` — in the repository, with the path in the evidence column
- `pending — assigned to the human` — needs a GitHub setting, an upload or a post
- `pending — needs code` — waiting on a task in the sprint

## Items

| Item | Status | Evidence or how-to |
| --- | --- | --- |
| LICENSE | done | `LICENSE` (MIT, copyright Adeel Ahmad); GitHub shows it in the sidebar |
| README banner and badges | done | `README.md` first screen, banner art in `docs/brand/readme-banner-dark.png` |
| About text (repo sidebar) | pending — assigned to the human | Sidebar ⚙ → Description. It still reads "more backends planned"; Restic is the only backend, so replace it with: "Time Machine-style restore for your Restic backups, in every directory. A .snapshot folder wherever you are; restoring a file is one cp." Set Website to `https://snapback.run` |
| Topics (12 set) | done | The live set matches the 12 under [Topics](#topics) below; read it back with `gh repo view adeelahmad/snapback --json repositoryTopics`, listed here in `docs/launch/checklist.md` |
| Social preview image | pending — assigned to the human | Settings → General → Social preview → Edit, upload `docs/brand/github-social-dark.png` (1280×640 PNG; GitHub rejects SVG here). This is the card every share on X, Slack, Discord, Hacker News and LinkedIn renders |
| Discussions enabled | done | Discussions are on; the issue chooser routes questions there, see `.github/ISSUE_TEMPLATE/config.yml` |
| Repository pinned on the profile | pending — assigned to the human | Profile → Customize your pins → tick snapback. The profile is where the author link from any post lands |
| First release with assets | done | v1.4.1, cut by `.github/workflows/release.yml`; the history is in `CHANGELOG.md` |
| Installer one-liner live | done | `install.sh`, published to `https://snapback.run/install.sh` by `.github/workflows/docs.yml` |
| CONTRIBUTING | done | `CONTRIBUTING.md` — build, the gate matrix, commit format |
| SECURITY.md | done | `SECURITY.md` — private advisory route and the disclosure window |
| Issue templates: bug and feature | done | `.github/ISSUE_TEMPLATE/bug.yml`, `.github/ISSUE_TEMPLATE/feature.yml` |
| Issue template: good first issue | done | `.github/ISSUE_TEMPLATE/good-first-issue.md` — what, where in the code, how to verify |
| Demo GIF in the README | pending — needs code | S5-19 records `ls .snapshot/` → `cp` → the web UI in about 15 seconds and places it under Quick start. For a command-line tool this is the single biggest conversion lever on the page |
| First posts | pending — assigned to the human | Hacker News ("Show HN"), r/selfhosted, r/DataHoarder, r/homelab, Lobsters and the Restic forum. Lead with the line below, never the feature list |
| Star-window timing | pending — assigned to the human | Post the first one early in the US morning on a weekday, and put the rest out within the same 24 hours so the stars land in one window. Be at the keyboard for the two hours after the Hacker News post: answering every comment fast is most of the difference between page two and the front page |

## The line to lead with

Every post opens with the restore itself, not with an architecture description:

```console
$ cp .snapshot/latest/report.docx .
```

One sentence of setup before it ("your backups show up as a read-only `.snapshot` folder
in the directory you are already in"), then the link. Anything longer buries it.

## Topics

```
backup  restore  restic  snapshot  fuse  time-machine  cli  golang  self-hosted  homelab  linux  macos
```

## Wording rules for every launch surface

- lowercase `snapback` everywhere, including the release title
- Restic is the only backend snapback supports; no other backend gets named or promised
- read-only: snapback never writes to a Restic repository, and the posts say so
