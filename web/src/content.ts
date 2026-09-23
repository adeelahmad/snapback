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
  pitch: string;
  installCommand: string;
  status: string;
}

export interface HowItWorksContent {
  heading: string;
  label: string;
  terminalLines: string[];
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
  pitch: 'snapback puts Time Machine-style restore inside the directories your Restic backups cover.',
  installCommand: 'curl -fsSL https://snapback.run/install.sh | sh',
  status: 'snapback v0.1 is early, pre-release software for Linux.',
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

export interface ListContent {
  heading: string;
  items: string[];
}

export const limits: ListContent = {
  heading: 'What snapback does not do',
  items: [
    'No backup scheduling. Run Restic on your own schedule.',
    'No retention policy. Pruning stays with Restic.',
    'No file-content cache.',
    'snapback snap is the one command that adds a snapshot to the repository, and only when you run it. snapback never deletes, prunes or rewrites repository data.',
    'No Windows support.',
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
