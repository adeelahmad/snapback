---
type: intake
story: C3
sprint: 2
---

# C3 intake — snapback.run React site in the snapback design system

Human request (2026-09-22T03:24:37Z): "add tasks to use this style and react pages site not just docs!" with a link to the user's own design system artifact (https://claude.ai/artifact/Va7MT8N7ELNW7aXU6XQP6i, title "snapback").

## Intent
1. **Goal:** snapback.run serves a React marketing/landing site built on the snapback design system, with the existing MkDocs docs still published (under /docs/) and install.sh at /install.sh.
2. **Users:** people who trust rclone/Restic/Syncthing-class tools; skeptical, technical.
3. **Must:** use the design system exactly — tokens (colour themes light/dark, IBM Plex Sans/Mono + JetBrains Mono, spacing, radius, shadow-glass), Liquid Glass containers over tinted radial grounds, flat mark/wordmark from the real SVG assets, man-page voice (lowercase "snapback", sentence case, no emoji, no exclamation marks, show the command first, state limitations). Honest content: only claim what exists (SPEC.md wording rules; the product today is the Stage 0 skeleton + Stage 1 compatibility evidence).
4. **Must not:** invent features; hats/shields/padlocks/clouds/drives/clocks/refresh arrows as icons; gradients (except the app icon); AI tropes; mix hues in the mark; pull fonts/scripts from CDNs at runtime (self-host fonts from the design system).
5. **Done when:** a static build (GitHub Pages) at snapback.run with landing page (hero with pitch + one command, how it works: .snapshot in every directory / cp to restore, what it does not do, install, status/roadmap honest to SPEC §22, footer links docs/GitHub), light+dark themes via prefers-color-scheme, WCAG AA contrast per token notes, favicon/app icon/OG meta from the assets; docs at /docs/; install.sh at /install.sh; CI builds and tests the site; gates green.

## Inputs (already downloaded, read-only)
- Design system files: /private/tmp/claude-501/-Users-adeelahmad-work-snapback/3a548722-26e5-4d22-8251-dbadece5e939/scratchpad/artifact-files/e75f5781-2fc6-4585-b3b2-227c107dbcbb/project/{README.md,tokens.json,design-system.json} and /private/tmp/claude-501/-Users-adeelahmad-work-snapback/3a548722-26e5-4d22-8251-dbadece5e939/scratchpad/design-system/project/{10-logo.md,20-assets.md,30-terminal.md,fonts/*.woff2,components/Cover/preview.html,components/SocialPreview/preview.html}.
- SVG/PNG assets (logos, icons, social) are in the artifact's asset store; the orchestrator will fetch them into the scratchpad before the asset task — plan a task that copies them from a given local folder into the site.

## Constraints from the repo
- Depends on C1 (snapback.run domain, docs.yml publishes install.sh) — C3 builds on C1's merged state; mkdocs moves to /docs/ (site_url https://snapback.run/docs/), so C1's tests on site_url/README docs link must be updated in C3.
- Toolchain pinning (standards.md): pin Node and every npm dependency exactly (lockfile committed); pinned actions by released version; no `latest`.
- Human primary stack: Next.js/React/Tailwind. Choose the simplest static option that fits GitHub Pages (Vite + React + TypeScript, or Next.js static export) and justify it.
