import { existsSync, readFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import { describe, expect, it } from 'vitest';
import { renderToStaticMarkup } from 'react-dom/server';
import { hero, howItWorks, restoreCompare, waysToRestore } from './content';
import { Header } from './sections/Header';
import { Hero } from './sections/Hero';
import { HowItWorks } from './sections/HowItWorks';
import { RestoreCompare } from './sections/RestoreCompare';
import { TwoProblems } from './sections/TwoProblems';
import { WaysToRestore } from './sections/WaysToRestore';

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

// textWithSpaces replaces each tag with a space before decoding entities, so
// text from adjacent elements never runs together across a tag boundary.
function textWithSpaces(markup: string): string {
  return decodeEntities(markup.replace(/<[^>]*>/g, ' '))
    .replace(/\s+/g, ' ')
    .trim();
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

// mediaBlock returns the text of the first `@media <query> {…}` block,
// matching braces by depth, or '' when there is none.
function mediaBlock(css: string, query: string): string {
  const marker = `@media ${query}`;
  const markerStart = css.indexOf(marker);
  if (markerStart === -1) return '';
  const braceStart = css.indexOf('{', markerStart);
  if (braceStart === -1) return '';

  let depth = 1;
  let i = braceStart + 1;
  while (i < css.length && depth > 0) {
    if (css[i] === '{') depth++;
    else if (css[i] === '}') depth--;
    i++;
  }
  if (depth !== 0) return '';
  return css.slice(braceStart + 1, i - 1);
}

// ruleBody returns the concatenated declarations of every rule, including
// rules inside media blocks, whose comma-separated selector list has an item
// equal to `selector` after trimming.
function ruleBody(css: string, selector: string): string {
  let body = '';
  for (const m of css.matchAll(/([^{}]+)\{([^{}]*)\}/g)) {
    const selectors = m[1].split(',').map((s) => s.trim());
    if (selectors.includes(selector)) {
      body += m[2];
    }
  }
  return body;
}

describe('sections', () => {
  it('heroShowsTheOneCommand', () => {
    const markup = renderToStaticMarkup(<Hero />);

    const got = countOccurrences(renderedText(markup), installCommand);
    expect(got, `occurrences of the install command in hero text`).toBe(1);
    const inCode = codeBlocks(markup).filter((b) => b.includes(installCommand));
    expect(inCode.length, `install command inside <code> or <pre>`).toBe(1);
  });

  it('heroSplitsBackupAndRestore', () => {
    const markup = renderToStaticMarkup(<Hero />);
    const text = textWithSpaces(markup);
    const sentences = text.split(/(?<=[.!?])\s+/);

    expect(
      sentences.some((s) => /\bbackup\b/i.test(s) && s.includes('Restic')),
      `a sentence naming backup and Restic in ${JSON.stringify(sentences)}`,
    ).toBe(true);
    expect(
      sentences.some(
        (s) => s.includes('snapback') && /\brestore\b/i.test(s) && !/backup tool/i.test(s),
      ),
      `a sentence naming snapback and restore, not backup tool, in ${JSON.stringify(sentences)}`,
    ).toBe(true);

    const h1Match = markup.match(/<h1[^>]*>([\s\S]*?)<\/h1>/);
    expect(h1Match, '<h1> present').toBeTruthy();
    const h1Text = decodeEntities((h1Match as RegExpMatchArray)[1].replace(/<[^>]*>/g, ''));
    expect(h1Text).toBe(hero.pitch);
    expect(h1Text).toContain('snapback');
    expect(h1Text).toContain('as easy as cp');
  });

  it('heroCallsSnapbackARestoreTool', () => {
    const markup = renderToStaticMarkup(<Hero />);

    const eyebrowMatch = markup.match(/<p class="hero__eyebrow">([\s\S]*?)<\/p>/);
    expect(eyebrowMatch, '<p class="hero__eyebrow"> present').toBeTruthy();
    const eyebrowText = decodeEntities(
      (eyebrowMatch as RegExpMatchArray)[1].replace(/<[^>]*>/g, ''),
    );
    expect(eyebrowText).toBe('A restore tool, not another backup tool');

    const eyebrowIndex = markup.indexOf('<p class="hero__eyebrow">');
    const h1Index = markup.indexOf('<h1');
    expect(h1Index, '<h1> present').toBeGreaterThan(-1);
    expect(eyebrowIndex, 'hero__eyebrow before <h1>').toBeLessThan(h1Index);

    const got = countOccurrences(renderedText(markup), installCommand);
    expect(got, `occurrences of the install command in hero text`).toBe(1);
  });

  it('heroStatesTheShippedRelease', () => {
    const text = renderedText(renderToStaticMarkup(<Hero />));

    expect(text).toContain('v0.1');
    expect(text).toContain('pre-release');
    expect(text).toContain('Restic');
    expect(text).toMatch(/\blinux\b/i);
    expect(text).not.toMatch(/stage 0|not yet|only command|is building/i);
  });

  it('heroShowsBackupAndRestorePanes', () => {
    const markup = renderToStaticMarkup(<Hero />);

    expect(hero.demo.length, 'hero.demo pane count').toBe(2);

    const [backupPane, restorePane] = hero.demo;
    expect(backupPane.label).toMatch(/backup/i);
    expect(backupPane.label).toMatch(/restic/i);
    expect(
      backupPane.lines.some((l) => l.text.startsWith('restic backup ')),
      `a backup pane line starting "restic backup " in ${JSON.stringify(backupPane.lines)}`,
    ).toBe(true);

    expect(restorePane.label).toMatch(/restore/i);
    expect(restorePane.label).toMatch(/snapback/i);
    expect(
      restorePane.lines.some((l) => l.text.startsWith('cp .snapshot/latest/')),
      `a restore pane line starting "cp .snapshot/latest/" in ${JSON.stringify(restorePane.lines)}`,
    ).toBe(true);

    for (const pane of hero.demo) {
      for (const line of pane.lines) {
        if (/\d{4}-\d{2}-\d{2}/.test(line.text)) {
          expect(
            line.text,
            `line with a date uses the shipped alias format: ${JSON.stringify(line.text)}`,
          ).toMatch(/\b\d{4}-\d{2}-\d{2}_\d{4}Z\b/);
        }
      }
    }

    const textIndex = markup.indexOf('<div class="hero__text">');
    const h1Index = markup.indexOf('<h1');
    const demoIndex = markup.indexOf('<div class="hero__demo">');
    expect(textIndex, 'hero__text present').toBeGreaterThan(-1);
    expect(h1Index, '<h1> present').toBeGreaterThan(-1);
    expect(demoIndex, 'hero__demo present').toBeGreaterThan(-1);
    expect(h1Index, '<h1> inside hero__text').toBeGreaterThan(textIndex);
    expect(demoIndex, 'hero__demo follows hero__text').toBeGreaterThan(h1Index);

    const captions = [
      ...markup.matchAll(/<figcaption class="terminal__label">([\s\S]*?)<\/figcaption>/g),
    ];
    expect(captions.length, 'hero demo figcaption count').toBe(2);

    const got = countOccurrences(renderedText(markup), installCommand);
    expect(got, `occurrences of the install command in hero text`).toBe(1);
  });

  it('stylesheetPutsHeroDemoBesideText', () => {
    const css = readStylesheet();

    expect(ruleBody(css, '.hero')).toContain('display: grid');

    const wide = mediaBlock(css, '(min-width: 960px)');
    expect(wide, '@media (min-width: 960px) block present').not.toBe('');
    expect(ruleBody(wide, '.hero')).toContain('grid-template-columns');

    expect(ruleBody(css, '.hero__lede')).toContain('max-width');
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
    expect(text).not.toMatch(/planned|not yet|\bwill\b/i);

    const aliasMatches = [...text.matchAll(/\b\d{4}-\d{2}-\d{2}_\d{4}Z\b/g)];
    expect(aliasMatches.length, `occurrences of the shipped alias format in ${JSON.stringify(lines)}`).toBeGreaterThanOrEqual(2);
    expect(text).toContain('latest');
    expect(lines.some((l) => /T\d{2}:\d{2}:\d{2}Z/.test(l)), `no line uses the old timestamp format in ${JSON.stringify(lines)}`).toBe(false);

    expect(text).toContain('How it works');
  });

  it('howItWorksRendersMechanismDetailsAfterTheTerminal', () => {
    const markup = renderToStaticMarkup(<HowItWorks />);

    const detailTexts = [...markup.matchAll(/<p class="how__detail">([\s\S]*?)<\/p>/g)].map((m) =>
      decodeEntities(m[1].replace(/<[^>]*>/g, '')),
    );
    expect(detailTexts).toEqual(howItWorks.details);
    expect(detailTexts).toEqual([
      'The only change snapback makes to a live directory is one managed .snapshot symlink.',
      'It points into a read-only FUSE catalog that lists the Restic snapshots containing that directory, plus a latest alias, and reads files from the repository only when you open them.',
      'While the daemon runs, a new snapshot can take up to about a minute to appear under .snapshot; snapback refresh asks it to reload sooner.',
    ]);

    const terminalIndex = markup.indexOf('class="glass terminal"');
    const firstDetailIndex = markup.indexOf('<p class="how__detail">');
    expect(terminalIndex, 'terminal present').toBeGreaterThan(-1);
    expect(firstDetailIndex, 'how__detail present').toBeGreaterThan(-1);
    expect(firstDetailIndex, 'details render after the terminal').toBeGreaterThan(terminalIndex);
  });

  it('twoProblemsNamesResticForBackupAndSnapbackForRestore', () => {
    const markup = renderToStaticMarkup(<TwoProblems />);

    expect(
      markup.startsWith('<section id="two-problems"'),
      'root section has id="two-problems"',
    ).toBe(true);

    const articles = tags(markup, 'article').map((a) => attr(a, 'class'));
    expect(articles).toEqual(['problem problem--backup', 'problem problem--restore']);

    const h3Texts = [...markup.matchAll(/<h3[^>]*>([\s\S]*?)<\/h3>/g)].map((m) => rawText(m[1]));
    expect(h3Texts).toEqual(['Backup', 'Restore']);

    const tagTexts = [...markup.matchAll(/<span class="problem__tag">([\s\S]*?)<\/span>/g)].map(
      (m) => rawText(m[1]),
    );
    expect(tagTexts).toEqual(['solved', 'an afterthought']);

    const backupArticleMatch = markup.match(
      /<article class="problem problem--backup">([\s\S]*?)<\/article>/,
    );
    expect(backupArticleMatch, 'backup article markup present').toBeTruthy();
    const backupText = rawText((backupArticleMatch as RegExpMatchArray)[1]);
    expect(backupText).toContain('Restic');
    expect(backupText).toContain('retention');

    const closingMatch = markup.match(/<p class="two-problems__closing">([\s\S]*?)<\/p>/);
    expect(closingMatch, 'closing paragraph present').toBeTruthy();
    expect(rawText((closingMatch as RegExpMatchArray)[1])).toContain(
      'snapback is only the restore half',
    );

    expect(renderedText(markup)).not.toMatch(/\b(borg|kopia|duplicati|time machine)\b/i);
  });

  it('restoreCompareShowsResticStepsAndOneCp', () => {
    const markup = renderToStaticMarkup(<RestoreCompare />);
    const lines = rawText(markup).split('\n').map((l) => l.trim());

    expect(
      markup.startsWith('<section id="restore-compare"'),
      'root section has id="restore-compare"',
    ).toBe(true);

    const captions = [...markup.matchAll(/<figcaption class="terminal__label">([\s\S]*?)<\/figcaption>/g)].map(
      (m) => rawText(m[1]),
    );
    expect(captions).toEqual(['restic only', 'with snapback']);

    expect(restoreCompare.before.lines.length, 'restoreCompare.before.lines length').toBeGreaterThanOrEqual(3);

    const beforeResticLines = restoreCompare.before.lines.filter((l) => l.text.startsWith('restic '));
    expect(beforeResticLines.length, `before lines starting "restic " in ${JSON.stringify(restoreCompare.before.lines)}`).toBeGreaterThanOrEqual(2);
    expect(
      beforeResticLines.some((l) => l.text.startsWith('restic restore ')),
      `a before line starting "restic restore " in ${JSON.stringify(restoreCompare.before.lines)}`,
    ).toBe(true);

    const afterCpLines = restoreCompare.after.lines.filter((l) => l.text.startsWith('cp '));
    expect(afterCpLines.length, `after lines starting "cp " in ${JSON.stringify(restoreCompare.after.lines)}`).toBe(1);
    expect(afterCpLines[0].text.startsWith('cp .snapshot/latest/')).toBe(true);

    for (const l of [...restoreCompare.before.lines, ...restoreCompare.after.lines]) {
      expect(['cmd', 'out']).toContain(l.kind);
    }

    expect(lines.some((l) => l.length > 0), 'rendered terminal lines present').toBe(true);
    expect(restoreCompare.intro).toContain('one cp');
  });

  it('waysToRestoreAreTheShippedSurfaces', () => {
    const markup = renderToStaticMarkup(<WaysToRestore />);

    expect(
      markup.startsWith('<section id="ways-to-restore"'),
      'root section has id="ways-to-restore"',
    ).toBe(true);

    const items = [...markup.matchAll(/<li class="way">([\s\S]*?)<\/li>/g)].map((m) =>
      rawText(m[1]),
    );
    expect(items.length, 'number of ways rendered').toBe(5);
    expect(items).toEqual(waysToRestore.items);

    const keys = ['cp', 'snapback open', 'snapback web', 'snapback doctor', 'snapback install service'];
    keys.forEach((key, i) => {
      expect(items[i], `item ${i} contains ${JSON.stringify(key)}`).toContain(key);
    });

    for (const item of waysToRestore.items) {
      expect(item, `item matches a banned surface: ${JSON.stringify(item)}`).not.toMatch(
        /\b(finder|macos|launchd|homebrew|apt|on-access|telemetry|setup)\b/i,
      );
    }
  });

  it('waysToRestoreWebUIMatchesTheHistoryPage', () => {
    const history = readFileSync(
      fileURLToPath(new URL('../../internal/webui/templates/history.html', import.meta.url)),
      'utf8',
    );
    expect(history).toContain('Restore copy next to original');

    const webItem = waysToRestore.items.find((item) => item.includes('snapback web'));
    expect(webItem, 'a waysToRestore item naming snapback web').toBeTruthy();
    expect(webItem).toContain('History');
    expect(webItem).toContain('next to the original');
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

  it('stylesheetDefinesResponsiveShell', () => {
    const css = readStylesheet();
    expect(css.length, 'web/src/styles/site.css is non-empty').toBeGreaterThan(0);

    expect(ruleBody(css, 'main')).toContain('max-width: 1120px');

    const desktop = mediaBlock(css, '(min-width: 768px)');
    expect(desktop, '@media (min-width: 768px) block present').not.toBe('');
    expect(ruleBody(desktop, 'main')).toContain('var(--space-8)');

    expect(ruleBody(css, 'h1')).toContain('clamp(');
    expect(ruleBody(css, 'h2')).toContain('clamp(');

    expect(ruleBody(css, 'main > section + section')).toContain(
      'border-top: 1px solid var(--line)',
    );
  });

  it('stylesheetKeepsOverflowInsideCodeBlocks', () => {
    const css = readStylesheet();
    expect(css.length, 'web/src/styles/site.css is non-empty').toBeGreaterThan(0);

    expect(ruleBody(css, 'pre')).toContain('overflow-x: auto');
    expect(ruleBody(css, 'img')).toContain('max-width: 100%');

    const overflowMasked = /overflow(-x)?:\s*(hidden|clip)/;
    for (const selector of ['html', 'body', 'main']) {
      expect(
        ruleBody(css, selector),
        `no ${selector} rule masks overflow`,
      ).not.toMatch(overflowMasked);
    }
  });

  it('stylesheetLaysProblemsSideBySide', () => {
    const css = readStylesheet();

    const desktop = mediaBlock(css, '(min-width: 768px)');
    expect(desktop, '@media (min-width: 768px) block present').not.toBe('');
    expect(ruleBody(desktop, '.problems')).toContain(
      'grid-template-columns: repeat(2, minmax(0, 1fr))',
    );

    expect(ruleBody(css, '.problem--restore')).toContain('var(--accent-soft)');
    expect(ruleBody(css, '.problem__tag')).toContain('var(--font-mono)');
    expect(ruleBody(css, '.problem__tag')).toContain('var(--radius-sm)');
  });

  it('stylesheetLaysTerminalsSideBySide', () => {
    const css = readStylesheet();

    const wide = mediaBlock(css, '(min-width: 960px)');
    expect(wide, '@media (min-width: 960px) block present').not.toBe('');
    expect(ruleBody(wide, '.compare')).toContain('grid-template-columns');

    expect(ruleBody(css, '.terminal__label')).toContain('var(--font-mono)');
    expect(ruleBody(css, '.terminal__label')).toContain('var(--ink-muted)');
    expect(ruleBody(css, '.terminal__out').length, '.terminal__out rule still exists').toBeGreaterThan(0);
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
