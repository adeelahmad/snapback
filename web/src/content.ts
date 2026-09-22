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
