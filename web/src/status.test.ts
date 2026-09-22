import { readFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import { describe, expect, it } from 'vitest';
import { stages } from './content';

// specStageNames returns the stage names from the SPEC.md §22 table, in order.
function specStageNames(): string[] {
  const spec = readFileSync(fileURLToPath(new URL('../../SPEC.md', import.meta.url)), 'utf8');
  const start = spec.indexOf('\n## 22.');
  const end = spec.indexOf('\n## 23.', start);
  const section = start < 0 ? '' : spec.slice(start, end < 0 ? undefined : end);
  return [...section.matchAll(/^\|\s*\d+\.\s+([^|]+?)\s*\|/gm)].map((m) => m[1]);
}

describe('status', () => {
  it('stageNamesMatchSPECSection22', () => {
    const want = specStageNames();

    expect(want.length, `stage rows parsed from SPEC.md §22: ${JSON.stringify(want)}`).toBe(8);
    const got = stages.map((s) => s.name);
    expect(got, 'stages.map(s => s.name)').toEqual(want);
  });

  it('onlyStage0IsDone', () => {
    expect(stages[0]?.state, 'stages[0].state').toBe('done');
    expect(stages[1]?.state, 'stages[1].state').toBe('in progress');
    const rest = stages.slice(2);
    expect(rest.length, 'stages 2-7 present').toBe(6);
    expect(
      rest.every((s) => s.state === 'planned'),
      `stages 2-7 states ${JSON.stringify(rest.map((s) => s.state))}`,
    ).toBe(true);
  });
});
