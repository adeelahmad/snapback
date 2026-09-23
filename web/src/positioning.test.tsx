import { readFileSync } from 'node:fs';
import { describe, expect, it } from 'vitest';
import { renderToStaticMarkup } from 'react-dom/server';
import App from './App';

const indexHTML = readFileSync(new URL('../index.html', import.meta.url), 'utf-8');

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

  it('TestMetaCarriesTheFraming', () => {
    const descriptionMatch = indexHTML.match(/<meta\s+name="description"\s+content="([^"]*)"/);
    const ogDescriptionMatch = indexHTML.match(
      /<meta\s+property="og:description"\s+content="([^"]*)"/,
    );
    const ldJsonMatch = indexHTML.match(
      /<script type="application\/ld\+json">([\s\S]*?)<\/script>/,
    );

    expect(descriptionMatch).not.toBeNull();
    expect(ogDescriptionMatch).not.toBeNull();
    expect(ldJsonMatch).not.toBeNull();

    const description = descriptionMatch![1];
    const ogDescription = ogDescriptionMatch![1];
    const ldJson = JSON.parse(ldJsonMatch![1]);
    const ldDescription = ldJson.description as string;

    expect(description).not.toBe('');
    expect(ogDescription).not.toBe('');
    expect(ldDescription).not.toBe('');
    expect(ogDescription).toBe(description);
    expect(ldDescription).toBe(description);

    const sentences = description.split(/(?<=[.!?])\s+/);
    expect(sentences[0]).toMatch(/\bbackup\b/);
    expect(sentences[0]).toContain('Restic');
    expect(sentences[1]).toContain('snapback');
    expect(sentences[1]).toContain('restore');

    expect(description.length).toBeGreaterThanOrEqual(50);
    expect(description.length).toBeLessThanOrEqual(160);

    expect(description).not.toMatch(/every directory|any directory/i);
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
