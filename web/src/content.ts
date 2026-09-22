// SUB-AGENT-TODO: fill every visible string for Header, Hero and HowItWorks per tasks.md § T4
// (lowercase `snapback`, sentence case, no `!`, planned features never in present tense).

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
  lockupAlt: '',
  lockupLight: '',
  lockupDark: '',
  links: [],
};

export const hero: HeroContent = {
  pitch: '',
  installCommand: '',
  status: '',
};

export const howItWorks: HowItWorksContent = {
  heading: '',
  label: '',
  terminalLines: [],
};
