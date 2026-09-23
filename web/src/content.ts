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
  pitch: 'snapback is building Time Machine-style restore for Restic backups, in every directory.',
  installCommand: 'curl -fsSL https://snapback.run/install.sh | sh',
  status:
    'Today snapback is the stage 0 skeleton, whose only command is snapback version, plus stage 1 compatibility evidence. The .snapshot view is not yet built.',
};

export const howItWorks: HowItWorksContent = {
  heading: 'How it will work',
  label: 'Planned behaviour, not yet built.',
  terminalLines: [
    'ls -a',
    '.  ..  .snapshot  report.docx',
    'ls .snapshot/',
    '2026-09-20T09:00:00Z  2026-09-21T09:00:00Z  latest',
    'cp .snapshot/2026-09-21T09:00:00Z/report.docx ./report.docx',
  ],
};

export type StageState = 'done' | 'in progress' | 'planned';

export interface Stage {
  name: string;
  state: StageState;
}

export const stages: Stage[] = [
  { name: 'Scaffolding', state: 'done' },
  { name: 'Compatibility milestone', state: 'in progress' },
  { name: 'Core vertical slice', state: 'planned' },
  { name: 'Reliable background operation', state: 'planned' },
  { name: 'Web UI and services', state: 'planned' },
  { name: 'macOS proof', state: 'planned' },
  { name: 'On-access mode', state: 'planned' },
  { name: 'Release', state: 'planned' },
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
