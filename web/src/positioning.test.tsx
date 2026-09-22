import { describe, expect, it } from 'vitest';
import { renderToStaticMarkup } from 'react-dom/server';
import App from './App';

const plannedBackends = ['Borg', 'Kopia', 'ZFS', 'Btrfs'];

function renderedText(markup: string): string {
  return markup
    .replace(/<[^>]*>/g, ' ')
    .replace(/&amp;/g, '&')
    .replace(/&#x27;|&#39;/g, "'")
    .replace(/\s+/g, ' ')
    .trim();
}

describe('positioning', () => {
  it('TestSupportsResticToday', () => {
    const text = renderedText(renderToStaticMarkup(<App />));

    expect(text).toContain('Restic today');
  });

  it('TestRoadmapListsPlannedBackends', () => {
    const text = renderedText(renderToStaticMarkup(<App />));
    const sentences = text.split(/(?<=[.!?])\s+/);

    const roadmap = sentences.filter(
      (s) => /\bplanned\b/i.test(s) && plannedBackends.every((b) => s.includes(b)),
    );
    expect(roadmap, `sentences naming ${plannedBackends.join(', ')} as planned`).not.toEqual([]);
  });

  it('TestNoBackendClaimedToWork', () => {
    const text = renderedText(renderToStaticMarkup(<App />));
    const sentences = text.split(/(?<=[.!?])\s+/);

    const named = sentences.filter((s) => plannedBackends.some((b) => s.includes(b)));
    expect(named, 'sentences naming a roadmap backend').not.toEqual([]);
    for (const s of named) {
      expect(s, 'roadmap backend sentence').not.toMatch(/\b(works|available now|now supports)\b/i);
    }
  });
});
