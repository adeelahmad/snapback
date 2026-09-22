// agentic:shim
// Compile shim for S2-12 T4. Deliberately wrong bodies so sections.test.tsx
// fails by assertion. The scaffolder deletes this file and points the test
// imports at ./sections/{Header,Hero,HowItWorks}.
import { createElement } from 'react';

export function Header() {
  return createElement('header', null, 'shim');
}

export function Hero() {
  return createElement('section', null, 'shim');
}

export function HowItWorks() {
  return createElement('section', null, 'shim');
}
