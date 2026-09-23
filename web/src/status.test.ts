import { readFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import { describe, expect, it } from 'vitest';
import { releases } from './content';

// releaseNamesFromSPEC returns the bold release names from the SPEC.md §22.1
// table, in order.
function releaseNamesFromSPEC(): string[] {
  const spec = readFileSync(fileURLToPath(new URL('../../SPEC.md', import.meta.url)), 'utf8');
  const start = spec.indexOf('\n### 22.1');
  const end = spec.indexOf('\n## 23.', start);
  const section = start < 0 ? '' : spec.slice(start, end < 0 ? undefined : end);
  return [...section.matchAll(/^\|\s*\*\*([^*]+)\*\*\s*\|/gm)].map((m) => m[1]);
}

describe('status', () => {
  it('releaseNamesMatchSPECSection22_1', () => {
    const want = releaseNamesFromSPEC();

    expect(want.length, `release names parsed from SPEC.md §22.1: ${JSON.stringify(want)}`).toBe(3);
    const got = releases.map((r) => r.name);
    expect(got, 'releases.map(r => r.name)').toEqual(want);
  });

  it('onlyV01IsShipped', () => {
    expect(releases[0]?.state, 'releases[0].state').toBe('shipped');
    const rest = releases.slice(1);
    expect(rest.length, 'releases 1-2 present').toBe(2);
    expect(
      rest.every((r) => r.state === 'planned'),
      `releases 1-2 states ${JSON.stringify(rest.map((r) => r.state))}`,
    ).toBe(true);
  });

  it('shippedScopeMatchesAcceptanceReport', () => {
    const report = readFileSync(
      fileURLToPath(new URL('../../docs/reports/v0.1-acceptance.md', import.meta.url)),
      'utf8',
    );
    expect(report.length, 'docs/reports/v0.1-acceptance.md is non-empty').toBeGreaterThan(0);
    expect(report, 'docs/reports/v0.1-acceptance.md').toContain('is met by Acc 1 to 17');

    expect(releases[0]?.scope, 'releases[0].scope').toContain('Linux');
    expect(releases[0]?.scope, 'releases[0].scope').toContain('Acceptance items 1 to 17 passed on Linux');

    expect(releases.length, 'releases').toBeGreaterThan(0);
    for (const release of releases) {
      const matches = release.scope.match(/macos[^.]*\bsupported\b/gi) ?? [];
      for (const match of matches) {
        expect(/\b(not|follow-up)\b/i.test(release.scope), `${release.name} scope contains "${match}"`).toBe(true);
      }
    }
  });
});
