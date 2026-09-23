export interface NavLink {
  label: string;
  href: string;
}

export interface HeaderContent {
  lockupAlt: string;
  lockupLight: string;
  lockupDark: string;
  links: NavLink[];
}

export interface HeroContent {
  eyebrow: string;
  pitch: string;
  lede: string;
  installCommand: string;
  status: string;
}

export interface Problem {
  name: string;
  tag: string;
  body: string;
}

export interface TwoProblemsContent {
  heading: string;
  intro: string;
  backup: Problem;
  restore: Problem;
  closing: string;
}

export interface TerminalLine {
  kind: 'cmd' | 'out';
  text: string;
}

export interface TerminalPane {
  label: string;
  lines: TerminalLine[];
}

export interface RestoreCompareContent {
  heading: string;
  intro: string;
  before: TerminalPane;
  after: TerminalPane;
}

export interface HowItWorksContent {
  heading: string;
  label: string;
  terminalLines: string[];
  details: string[];
}

export const header: HeaderContent = {
  lockupAlt: 'snapback',
  lockupLight: '/brand/lockup-horizontal.svg',
  lockupDark: '/brand/lockup-horizontal-inverse.svg',
  links: [
    { label: 'docs', href: '/docs/' },
    { label: 'GitHub', href: 'https://github.com/adeelahmad/snapback' },
  ],
};

export const hero: HeroContent = {
  eyebrow: 'A restore tool, not another backup tool',
  pitch: 'Backup is a solved problem. snapback makes restore as easy as cp.',
  lede:
    'Restic keeps doing the backup: on your schedule, deduplicated and encrypted, under your retention rules. snapback does the restore: a read-only .snapshot entry inside the directories your backups cover, so getting a file back is as easy as it was in 2008.',
  installCommand: 'curl -fsSL https://snapback.run/install.sh | sh',
  status: 'snapback v0.1 is early, pre-release software for Linux.',
};

export const twoProblems: TwoProblemsContent = {
  heading: 'Two problems, not one',
  intro:
    'Backup tools are built for the day you back up. Restore is for the day you need a file back.',
  backup: {
    name: 'Backup',
    tag: 'solved',
    body: 'Restic takes deduplicated, encrypted snapshots whenever your cron job or timer runs it, and forgets and prunes them by your retention rules. snapback does not compete with it.',
  },
  restore: {
    name: 'Restore',
    tag: 'an afterthought',
    body: 'Every backup tool can restore. Few make it easy: open another program, find the snapshot and the path again, restore into a scratch directory, then move the file into place.',
  },
  closing:
    'snapback is only the restore half. It reads the Restic repository you already have and puts earlier versions of a file next to the file, as ordinary read-only directories any program can read.',
};

export const restoreCompare: RestoreCompareContent = {
  heading: 'Restore without a restore session',
  intro:
    'Getting report.docx back with Restic alone takes a snapshot ID, a path and a scratch directory. With snapback it is one cp.',
  before: {
    label: 'restic only',
    lines: [
      { kind: 'cmd', text: 'restic snapshots --path ~/Documents' },
      {
        kind: 'cmd',
        text: 'restic restore 4f2a9c1e --target /tmp/restore --include ~/Documents/report.docx',
      },
      { kind: 'cmd', text: 'cp /tmp/restore/home/you/Documents/report.docx ~/Documents/' },
    ],
  },
  after: {
    label: 'with snapback',
    lines: [
      { kind: 'cmd', text: 'cd ~/Documents' },
      { kind: 'cmd', text: 'cp .snapshot/latest/report.docx .' },
    ],
  },
};

export const howItWorks: HowItWorksContent = {
  heading: 'How it works',
  label:
    'Each directory your backups cover gets one read-only .snapshot entry: a folder per Restic snapshot that contains it, plus a latest alias.',
  terminalLines: [
    'ls -a',
    '.  ..  .snapshot  report.docx',
    'ls .snapshot/',
    '2026-09-21_0300Z  2026-09-22_0300Z  latest',
    'cp .snapshot/latest/report.docx .',
  ],
  details: [
    'The only change snapback makes to a live directory is one managed .snapshot symlink.',
    'It points into a read-only FUSE catalog that lists the Restic snapshots containing that directory, plus a latest alias, and reads files from the repository only when you open them.',
    'While the daemon runs, a new snapshot can take up to about a minute to appear under .snapshot; snapback refresh asks it to reload sooner.',
  ],
};

export const waysToRestore: ListContent = {
  heading: 'Ways to get a file back',
  items: [
    'Any program: cp, diff, an editor or a file manager reads .snapshot like any other directory.',
    "snapback open shows a directory's history in your file manager.",
    'snapback web serves a local web UI whose History view can restore a copy next to the original.',
    "snapback doctor checks the prerequisites and your repository's health.",
    'snapback install service runs the daemon as a systemd user service on Linux.',
  ],
};

export type ReleaseState = 'shipped' | 'planned';

export interface Release {
  name: string;
  state: ReleaseState;
  scope: string;
}

export const releases: Release[] = [
  {
    name: 'v0.1, first public release candidate',
    state: 'shipped',
    scope:
      'Linux: the .snapshot view with its latest alias, snapback snap, snapback seed and snapback link, the daemon, the local web UI, snapback doctor and the systemd user service. Acceptance items 1 to 17 passed on Linux.',
  },
  {
    name: 'Follow-up, separately proven',
    state: 'planned',
    scope:
      'macOS with macFUSE, launchd, the Finder companion and more package channels. Each ships only with its own acceptance evidence.',
  },
  {
    name: 'Experimental until acceptance proof',
    state: 'planned',
    scope:
      'On-access mode (Linux fanotify, macOS Endpoint Security), opt-in and labelled experimental until its acceptance test passes.',
  },
];

export interface EvidenceLink {
  lead: string;
  label: string;
  href: string;
}

export const statusEvidence: EvidenceLink = {
  lead: 'Item-by-item results are in',
  label: 'the v0.1 acceptance report',
  href: 'https://github.com/adeelahmad/snapback/blob/master/docs/reports/v0.1-acceptance.md',
};

export interface ListContent {
  heading: string;
  items: string[];
}

export const resticKeeps: ListContent = {
  heading: 'What Restic keeps doing',
  items: [
    'Scheduling. Run restic backup from cron or a systemd timer, as you do now.',
    'Retention and pruning. restic forget and restic prune stay with you; snapback never deletes, prunes or rewrites repository data.',
    'One exception, on request: snapback snap asks Restic to take one snapshot of a directory now, and only when you run it.',
  ],
};

export const limits: ListContent = {
  heading: 'What snapback does not do',
  items: [
    'No Windows support. FUSE is a hard requirement.',
    'No file-content cache. Files are read from the repository only when you open them.',
    'No live overlay or union of snapshot and working files.',
    'snapback supports Restic today; other backends are not yet supported.',
    'Borg, Kopia, ZFS and Btrfs snapshots are planned backends, not yet supported.',
  ],
};

export const install: ListContent = {
  heading: 'Install',
  items: [
    'The script downloads the latest release for your operating system and architecture and verifies it against the signed checksums.txt before it installs anything.',
    'snapback needs FUSE (fuse3 on Linux), the restic CLI and an existing Restic repository.',
    'Run snapback doctor afterwards to check the prerequisites and your repository.',
  ],
};

export const statusHeading = 'Status';

export const footerLinks: NavLink[] = [
  { label: 'docs', href: '/docs/' },
  { label: 'GitHub', href: 'https://github.com/adeelahmad/snapback' },
  { label: 'licence', href: 'https://github.com/adeelahmad/snapback/blob/master/LICENSE' },
];
