// agentic:shim
// Compile shim for S2-12 T5. Deliberately wrong bodies so copy.test.tsx and
// status.test.ts fail by assertion. The scaffolder deletes this file and points
// the test imports at ./sections/{Limits,Install,Footer} and ./content.
import { createElement } from 'react';

export type StageState = 'done' | 'in progress' | 'planned';

export interface Stage {
  name: string;
  state: StageState;
}

export const stages: Stage[] = [{ name: 'shim', state: 'planned' }];

export function Limits() {
  return createElement('section', null, 'shim');
}

export function Install() {
  return createElement('section', null, 'shim');
}

export function Footer() {
  return createElement('footer', null, createElement('a', { href: 'http://shim' }, 'shim'));
}
