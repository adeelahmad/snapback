import { hero } from '../content';

export function Hero() {
  return (
    <section id="hero" className="hero">
      <p className="hero__eyebrow">{hero.eyebrow}</p>
      <h1>{hero.pitch}</h1>
      <p className="hero__lede">{hero.lede}</p>
      <div className="glass snippet">
        <pre>
          <code>{hero.installCommand}</code>
        </pre>
      </div>
      <p className="hero__status">{hero.status}</p>
    </section>
  );
}
