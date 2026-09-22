import { describe, expect, it } from 'vitest';
import { renderToStaticMarkup } from 'react-dom/server';
import App from './App';
import * as content from './content';
import { Footer, Install, Limits } from './zz_agentic_shim_t5';

const installCommand = 'curl -fsSL https://snapback.run/install.sh | sh';
const repoURL = 'https://github.com/adeelahmad/snapback';

const headingAllowlist = new Set([
  'Restic',
  'Time',
  'Machine',
  'GitHub',
  'Linux',
  'macOS',
  'FUSE',
  'SPEC',
]);

const bannedClaims = [
  'production-ready',
  'cross-platform',
  'static',
  'finder-integrated',
  'available now',
  'works today',
  'now supports',
  'first-of-its-kind',
  'borg',
  'kopia',
  'duplicity',
  'duplicati',
  'tarsnap',
  'rustic',
  'multi-backend',
];

function decodeEntities(s: string): string {
  return s
    .replace(/&lt;/g, '<')
    .replace(/&gt;/g, '>')
    .replace(/&quot;/g, '"')
    .replace(/&#x27;|&#39;/g, "'")
    .replace(/&nbsp;/g, ' ')
    .replace(/&amp;/g, '&');
}

function renderedText(markup: string): string {
  return decodeEntities(markup.replace(/<[^>]*>/g, ' '))
    .replace(/\s+/g, ' ')
    .trim();
}

function tags(markup: string, name: string): string[] {
  return [...markup.matchAll(new RegExp(`<${name}\\b[^>]*>`, 'g'))].map((m) => m[0]);
}

function attr(tag: string, name: string): string | undefined {
  const m = tag.match(new RegExp(`\\s${name}="([^"]*)"`));
  return m ? decodeEntities(m[1]) : undefined;
}

function headings(markup: string): string[] {
  return [...markup.matchAll(/<(h[1-3])\b[^>]*>([\s\S]*?)<\/\1>/g)].map((m) =>
    renderedText(m[2]),
  );
}

// contentStrings returns every string value reachable from the content module.
function contentStrings(value: unknown): string[] {
  if (typeof value === 'string') {
    return [value];
  }
  if (Array.isArray(value)) {
    return value.flatMap(contentStrings);
  }
  if (value !== null && typeof value === 'object') {
    return Object.values(value).flatMap(contentStrings);
  }
  return [];
}

function countOccurrences(haystack: string, needle: string): number {
  return haystack.split(needle).length - 1;
}

describe('copy', () => {
  it('TestSectionOrder', () => {
    const markup = renderToStaticMarkup(<App />);

    const got = tags(markup, 'section').map((s) => attr(s, 'id') ?? '');
    expect(got, '<section id> values in <App/>').toEqual([
      'hero',
      'how-it-works',
      'limits',
      'install',
      'status',
    ]);
    const lastSection = markup.lastIndexOf('</section>');
    const footer = markup.indexOf('<footer', lastSection);
    expect(footer, '<footer> after the last </section>').toBeGreaterThan(lastSection);
  });

  it('TestVoiceLint', () => {
    const markup = renderToStaticMarkup(<App />);
    const text = renderedText(markup);

    expect(text.length, 'rendered <App/> text length').toBeGreaterThan(200);
    const corpus = [text, ...contentStrings(content)];
    for (const s of corpus) {
      expect(s, 'no emoji').not.toMatch(/\p{Extended_Pictographic}/u);
      expect(s, 'no exclamation mark').not.toContain('!');
      expect(s, 'no capitalised Snapback').not.toContain('Snapback');
      expect(s, 'no "snap back"').not.toMatch(/snap back/i);
    }
    for (const h of headings(markup)) {
      const words = h.match(/\p{L}[\p{L}\p{N}]*/gu) ?? [];
      const titled = words
        .slice(1)
        .filter((w) => /^\p{Lu}/u.test(w) && !headingAllowlist.has(w));
      expect(titled, `Title Case words in heading ${JSON.stringify(h)}`).toEqual([]);
    }
  });

  it('TestHonesty', () => {
    const text = renderedText(renderToStaticMarkup(<App />));
    const lower = text.toLowerCase();

    for (const claim of bannedClaims) {
      expect(lower.includes(claim), `rendered text contains ${JSON.stringify(claim)}`).toBe(false);
    }
    const sentences = text.split(/(?<=[.!?])\s+/);
    for (const s of sentences) {
      if (!s.includes('.snapshot')) {
        continue;
      }
      const claimsNow = /\b(today|now)\b/i.test(s) && !/not yet/i.test(s);
      expect(claimsNow, `sentence claims .snapshot works now: ${JSON.stringify(s)}`).toBe(false);
    }
    expect(text).toContain('not yet');
    expect(text).toContain('Restic');
  });

  it('TestLimitsSection', () => {
    const text = renderedText(renderToStaticMarkup(<Limits />));

    expect(text).toMatch(/schedul/i);
    expect(text).toMatch(/retention/i);
    expect(text).toContain('Windows');
    expect(text, '"never writes" near "Restic repository"').toMatch(
      /never writes[^.]{0,60}Restic repository/i,
    );
  });

  it('TestInstallSection', () => {
    const text = renderedText(renderToStaticMarkup(<Install />));

    expect(countOccurrences(text, installCommand), 'occurrences of the install command').toBe(1);
    expect(text).toContain('snapback version');
    expect(text).toMatch(/checksum/i);
  });

  it('TestFooterLinks', () => {
    const anchors = tags(renderToStaticMarkup(<Footer />), 'a');

    const hrefs = anchors.map((a) => attr(a, 'href'));
    expect(hrefs).toContain('/docs/');
    expect(hrefs).toContain(repoURL);
    for (const a of anchors) {
      if (!(attr(a, 'href') ?? '').startsWith('http')) {
        continue;
      }
      expect(attr(a, 'rel') ?? '', `rel of ${a}`).toContain('noopener');
    }
  });
});
