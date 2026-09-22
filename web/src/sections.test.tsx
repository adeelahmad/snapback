import { existsSync, readFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import { describe, expect, it } from 'vitest';
import { renderToStaticMarkup } from 'react-dom/server';
import { Header, Hero, HowItWorks } from './zz_shim_t4';

const installCommand = 'curl -fsSL https://snapback.run/install.sh | sh';
const repoURL = 'https://github.com/adeelahmad/snapback';

function decodeEntities(s: string): string {
  return s
    .replace(/&lt;/g, '<')
    .replace(/&gt;/g, '>')
    .replace(/&quot;/g, '"')
    .replace(/&#x27;|&#39;/g, "'")
    .replace(/&nbsp;/g, ' ')
    .replace(/&amp;/g, '&');
}

// rawText strips tags and decodes entities but keeps newlines.
function rawText(markup: string): string {
  return decodeEntities(markup.replace(/<[^>]*>/g, ''));
}

function renderedText(markup: string): string {
  return rawText(markup).replace(/\s+/g, ' ').trim();
}

function countOccurrences(haystack: string, needle: string): number {
  return haystack.split(needle).length - 1;
}

function codeBlocks(markup: string): string[] {
  const blocks: string[] = [];
  for (const m of markup.matchAll(/<(code|pre)\b[^>]*>([\s\S]*?)<\/\1>/g)) {
    blocks.push(rawText(m[2]));
  }
  return blocks;
}

function tags(markup: string, name: string): string[] {
  return [...markup.matchAll(new RegExp(`<${name}\\b[^>]*>`, 'g'))].map((m) => m[0]);
}

function attr(tag: string, name: string): string | undefined {
  const m = tag.match(new RegExp(`\\s${name}="([^"]*)"`));
  return m ? decodeEntities(m[1]) : undefined;
}

function readStylesheet(): string {
  const path = fileURLToPath(new URL('./styles/site.css', import.meta.url));
  return existsSync(path) ? readFileSync(path, 'utf8') : '';
}

describe('sections', () => {
  it('heroShowsTheOneCommand', () => {
    const markup = renderToStaticMarkup(<Hero />);

    const got = countOccurrences(renderedText(markup), installCommand);
    expect(got, `occurrences of the install command in hero text`).toBe(1);
    const inCode = codeBlocks(markup).filter((b) => b.includes(installCommand));
    expect(inCode.length, `install command inside <code> or <pre>`).toBe(1);
  });

  it('heroStatesWhatExistsToday', () => {
    const text = renderedText(renderToStaticMarkup(<Hero />));

    expect(text).toMatch(/stage 0/i);
    expect(text).toMatch(/not yet/i);
  });

  it('howItWorksShowsSnapshotAndCp', () => {
    const markup = renderToStaticMarkup(<HowItWorks />);
    const text = renderedText(markup);

    expect(text).toContain('.snapshot');
    const lines = rawText(markup).split('\n').map((l) => l.trim());
    expect(
      lines.some((l) => l.startsWith('cp .snapshot/')),
      `a line starting "cp .snapshot/" in ${JSON.stringify(lines)}`,
    ).toBe(true);
    expect(text).toMatch(/planned/i);
  });

  it('headerLinksDocsAndGitHub', () => {
    const markup = renderToStaticMarkup(<Header />);

    const hrefs = tags(markup, 'a').map((a) => attr(a, 'href'));
    expect(hrefs).toContain('/docs/');
    expect(hrefs).toContain(repoURL);
    const lockups = tags(markup, 'img').filter(
      (img) => attr(img, 'src') === '/brand/lockup-horizontal.svg',
    );
    expect(lockups.length, 'img with src /brand/lockup-horizontal.svg').toBeGreaterThan(0);
    for (const img of lockups) {
      expect(attr(img, 'alt')).toBe('snapback');
    }
  });

  it('stylesheetUsesTokensOnly', () => {
    const css = readStylesheet();

    expect(css.length, 'web/src/styles/site.css is non-empty').toBeGreaterThan(0);
    expect(css).not.toMatch(/#[0-9a-fA-F]{3,8}\b/);
    expect(css).not.toMatch(/rgb\(|hsl\(/);
    expect(css).not.toMatch(/linear-gradient/);
    expect(css).toMatch(/radial-gradient\([^;]*var\(--glass-tint-/);
    expect(css).toContain('var(--shadow-glass)');
  });

  it('stylesheetSetsNoBlueOrYellowText', () => {
    const css = readStylesheet();

    expect(css.length, 'web/src/styles/site.css is non-empty').toBeGreaterThan(0);
    expect(css).not.toMatch(/color:\s*var\(--(blue|blue-alt|yellow|yellow-text|red)\)/);
  });
});
