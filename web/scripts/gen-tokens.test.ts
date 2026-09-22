import { describe, expect, it } from 'vitest';
import { existsSync, readFileSync } from 'node:fs';
import { basename } from 'node:path';
import { fileURLToPath } from 'node:url';

type ThemeValue = string | { light: string; dark: string };

interface Tokens {
  color: { tokens: { name: string; value: ThemeValue }[] };
  type: { fonts: { family: string; file: string; weight: string }[] };
  spacing: { tokens: { name: string; value: string }[] };
  radius: { tokens: { name: string; value: string }[] };
  shadow: { tokens: { name: string; value: ThemeValue }[] };
}

type Generate = (tokens: Tokens) => string;

const DARK_MEDIA = '@media (prefers-color-scheme: dark)';

function repoFile(rel: string): string {
  const path = fileURLToPath(new URL(`../${rel}`, import.meta.url));
  expect(existsSync(path), `web/${rel} must exist`).toBe(true);
  return readFileSync(path, 'utf8');
}

function loadTokens(): Tokens {
  return JSON.parse(repoFile('tokens.json')) as Tokens;
}

async function loadGenerate(): Promise<Generate> {
  const url = new URL('./gen-tokens.mjs', import.meta.url);
  expect(existsSync(fileURLToPath(url)), 'web/scripts/gen-tokens.mjs must exist').toBe(true);
  const mod = (await import(/* @vite-ignore */ url.href)) as { generate?: unknown };
  expect(typeof mod.generate, 'gen-tokens.mjs must export generate').toBe('function');
  return mod.generate as Generate;
}

// A reference value such as "{blue}" becomes var(--blue); the other token
// carries the per-theme value.
function cssValue(v: string): string {
  const ref = v.match(/^\{([a-z0-9-]+)\}$/);
  return ref ? `var(--${ref[1]})` : v;
}

function themeValue(v: ThemeValue, theme: 'light' | 'dark'): string {
  return cssValue(typeof v === 'string' ? v : v[theme]);
}

async function generatedBlocks(): Promise<{ css: string; light: string; dark: string; tokens: Tokens }> {
  const tokens = loadTokens();
  const generate = await loadGenerate();
  const css = generate(tokens);
  const at = css.indexOf(DARK_MEDIA);
  expect(at, `output must contain ${DARK_MEDIA}`).toBeGreaterThan(-1);
  return { css, light: css.slice(0, at), dark: css.slice(at), tokens };
}

describe('gen-tokens', () => {
  it('emitsEveryColourTokenForBothThemes', async () => {
    const { light, dark, tokens } = await generatedBlocks();

    expect(tokens.color.tokens).toHaveLength(30);
    for (const { name, value } of tokens.color.tokens) {
      expect(light).toContain(`--${name}: ${themeValue(value, 'light')};`);
      expect(dark).toContain(`--${name}: ${themeValue(value, 'dark')};`);
    }
  });

  it('emitsSpacingRadiusAndShadowTokens', async () => {
    const { light, dark, tokens } = await generatedBlocks();

    const flat = [...tokens.spacing.tokens, ...tokens.radius.tokens];
    expect(flat.length).toBeGreaterThan(0);
    for (const { name, value } of flat) {
      expect(light).toContain(`--${name}: ${value};`);
    }
    expect(tokens.shadow.tokens.length).toBeGreaterThan(0);
    for (const { name, value } of tokens.shadow.tokens) {
      expect(light).toContain(`--${name}: ${themeValue(value, 'light')};`);
      expect(dark).toContain(`--${name}: ${themeValue(value, 'dark')};`);
    }
  });

  it('emitsOneFontFacePerFontFile', async () => {
    const { css, tokens } = await generatedBlocks();

    expect(tokens.type.fonts.length).toBeGreaterThan(0);
    const faces = [...css.matchAll(/@font-face\s*\{([^}]*)\}/g)].map((m) => m[1]);
    expect(css.match(/@font-face/g) ?? []).toHaveLength(tokens.type.fonts.length);
    for (const font of tokens.type.fonts) {
      const src = `url('/fonts/${basename(font.file)}') format('woff2')`;
      const face = faces.find((f) => f.includes(src));
      expect(face, `@font-face with ${src}`).toBeDefined();
      expect(face).toMatch(new RegExp(`font-weight:\\s*${font.weight}\\s*;`));
      expect(face).toMatch(/font-display:\s*swap\s*;/);
    }
  });

  it('declaresColorSchemeLightDark', async () => {
    const { light } = await generatedBlocks();

    const root = light.match(/:root\s*\{([^}]*)\}/);
    expect(root, ':root block before the dark media query').not.toBeNull();
    expect(root?.[1]).toContain('color-scheme: light dark;');
  });

  it('committedTokensCssMatchesTheGenerator', async () => {
    const tokens = loadTokens();
    const generate = await loadGenerate();

    const committed = repoFile('src/styles/tokens.css');
    expect(committed).toBe(generate(tokens));
  });

  it('isDeterministic', async () => {
    const tokens = loadTokens();
    const generate = await loadGenerate();
    const pristine = structuredClone(tokens);
    const input = structuredClone(tokens);

    const first = generate(input);
    const second = generate(input);

    expect(first.length).toBeGreaterThan(0);
    expect(second).toBe(first);
    expect(input).toEqual(pristine);
  });
});
