import { describe, expect, it } from 'vitest';
import { renderToStaticMarkup } from 'react-dom/server';
import App from './App';

describe('App', () => {
  it('rendersAMainLandmarkWithTheLowercaseName', () => {
    const markup = renderToStaticMarkup(<App />);

    expect(markup).toContain('<main');
    const h1 = markup.match(/<h1[^>]*>([\s\S]*?)<\/h1>/);
    expect(h1).not.toBeNull();
    const text = (h1?.[1] ?? '').replace(/<[^>]*>/g, '');
    expect(text).toContain('snapback');
  });
});
