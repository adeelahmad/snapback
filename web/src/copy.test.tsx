import { existsSync, readFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import { describe, expect, it } from 'vitest';
import { renderToStaticMarkup } from 'react-dom/server';
import App from './App';
import * as content from './content';
import { Footer } from './sections/Footer';
import { Install } from './sections/Install';
import { Limits } from './sections/Limits';
import { Status } from './sections/Status';

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
  'Borg',
  'Kopia',
  'ZFS',
  'Btrfs',
]);

// roadmapBackends may appear only in sentences that call them planned.
const roadmapBackends = /\b(borg|kopia|zfs|btrfs)\b/i;

const bannedClaims = [
  'production-ready',
  'cross-platform',
  'static',
  'finder-integrated',
  'available now',
  'works today',
  'now supports',
  'first-of-its-kind',
  'duplicity',
  'duplicati',
  'tarsnap',
  'rustic',
  'multi-backend',
  'never writes',
  'only reads',
  'snapback setup',
  'telemetry',
  'docker',
  'stage 0',
  'not yet built',
  'only command',
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
      'two-problems',
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
      expect(s, `.snapshot sentence claims future or unshipped: ${JSON.stringify(s)}`).not.toMatch(
        /\b(will|planned|not yet built)\b/i,
      );
      expect(s, `.snapshot sentence names macOS or Windows: ${JSON.stringify(s)}`).not.toMatch(
        /\b(macos|windows)\b/i,
      );
    }
    for (const s of sentences) {
      if (!roadmapBackends.test(s)) {
        continue;
      }
      expect(s, 'sentence naming a roadmap backend').toMatch(/\bplanned\b/i);
    }
    expect(text).toContain('not yet');
    expect(text).toContain('Restic');
    expect(text).toContain('Linux');
  });

  it('TestLimitsSection', () => {
    const text = renderedText(renderToStaticMarkup(<Limits />));

    expect(text).toMatch(/schedul/i);
    expect(text).toMatch(/retention/i);
    expect(text).toContain('Windows');
    expect(text, '"snapback snap" is the only command that adds a snapshot').toMatch(
      /snapback snap[^.]{0,80}only when you run it/i,
    );
    expect(text, 'never deletes or prunes repository data').toMatch(
      /never (deletes|prunes)[^.]{0,40}repository/i,
    );
    expect(text, 'no "never writes" or "only reads" claim').not.toMatch(
      /never writes|only reads/i,
    );
  });

  it('TestNoSentenceMakesSnapbackTheBackupTool', () => {
    const text = renderedText(renderToStaticMarkup(<App />));
    expect(text, 'rendered <App/> text mentions snapback').toContain('snapback');

    const sentences = text.split(/(?<=[.!?])\s+/);
    for (const s of sentences) {
      expect(
        s,
        `sentence makes snapback the backup tool: ${JSON.stringify(s)}`,
      ).not.toMatch(/\bsnapback (schedules|prunes|backs up|replaces|is a backup tool)\b/i);

      if (s.includes('snapback') && /backup tool/i.test(s)) {
        expect(
          s,
          `sentence names snapback and "backup tool" without "not": ${JSON.stringify(s)}`,
        ).toMatch(/\bnot\b/i);
      }

      expect(
        s,
        `sentence claims snapback replaces Restic: ${JSON.stringify(s)}`,
      ).not.toMatch(/\b(replaces|replacement for) restic\b/i);
    }
  });

  it('TestInstallSection', () => {
    const text = renderedText(renderToStaticMarkup(<Install />));

    expect(countOccurrences(text, installCommand), 'occurrences of the install command').toBe(1);
    expect(text).toContain('fuse3');
    expect(text).toContain('restic CLI');
    expect(text).toContain('Restic repository');
    expect(text).toContain('snapback doctor');
    expect(text).toMatch(/checksum/i);
    expect(text).not.toContain('snapback version');
    expect(text).not.toContain('only command');
  });

  it('TestNamedCommandsAreShipped', () => {
    const usage = readFileSync(fileURLToPath(new URL('../../docs-site/usage.md', import.meta.url)), 'utf8');
    const usageCommands = new Set(
      [...usage.matchAll(/^\| `snapback ([a-z][a-z-]*)` \|/gm)].map((m) => m[1]),
    );
    expect(usageCommands.size, 'commands parsed from docs-site/usage.md').toBeGreaterThanOrEqual(10);

    const proseWords = new Set(['puts', 'v0.1', 'never', 'supports', 'needs', 'makes', 'does', 'is']);
    const captures = contentStrings(content).flatMap((s) =>
      [...s.matchAll(/\bsnapback\s+([a-z][a-z0-9.-]*[a-z0-9])/g)].map((m) => m[1]),
    );
    const usageCaptures = captures.filter((c) => usageCommands.has(c));
    expect(usageCaptures.length, `captures naming a usage command: ${JSON.stringify(captures)}`).toBeGreaterThanOrEqual(1);
    for (const c of captures) {
      expect(
        usageCommands.has(c) || proseWords.has(c),
        `capture ${JSON.stringify(c)} is neither a usage command nor an allowed prose word`,
      ).toBe(true);
    }

    expect(usageCommands.has('setup'), 'usage table lists a setup command').toBe(false);
    expect(usageCommands.has("telemetry"), 'usage table lists a telemetry command').toBe(false);
  });

  it('TestStatusCitesAcceptanceEvidence', () => {
    const markup = renderToStaticMarkup(<Status />);
    const text = renderedText(markup);

    const anchors = tags(markup, 'a').filter((a) => (attr(a, 'href') ?? '').endsWith('/docs/reports/v0.1-acceptance.md'));
    expect(anchors.length, 'anchors linking to the v0.1 acceptance report').toBe(1);
    expect(attr(anchors[0], 'rel') ?? '', 'rel of the acceptance report link').toContain('noopener');

    const href = attr(anchors[0], 'href') ?? '';
    const repoPath = href.split('/blob/master/')[1];
    expect(repoPath, 'repo-relative path parsed from the evidence href').toBeTruthy();
    expect(
      existsSync(fileURLToPath(new URL(`../../${repoPath}`, import.meta.url))),
      `${repoPath} exists on disk`,
    ).toBe(true);

    expect(text).toContain('acceptance report');
  });

  it('TestStatusMatchesReadme', () => {
    const readme = readFileSync(fileURLToPath(new URL('../../README.md', import.meta.url)), 'utf8');
    const start = readme.indexOf('\n## Status\n');
    const end = readme.indexOf('\n## ', start + 1);
    const section = start < 0 ? '' : readme.slice(start, end < 0 ? undefined : end);

    expect(section.length, 'README.md ## Status section is non-empty').toBeGreaterThan(0);
    expect(section).toContain('v0.1');
    expect(section).toContain('pre-release');

    expect(content.hero.status).toContain('v0.1');
    expect(content.hero.status).toContain('pre-release');
    expect(content.hero.status).toMatch(/\blinux\b/i);
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

  it('TestFooterAnalyticsNote', () => {
    const text = renderedText(renderToStaticMarkup(<Footer />));

    expect(text, 'footer privacy note').toContain('Google Analytics');
  });
});
