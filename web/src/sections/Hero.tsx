import { hero } from '../content';

export function Hero() {
  return (
    <section id="hero" className="hero">
      <h1>{hero.pitch}</h1>
      <div className="glass snippet">
        <pre>
          <code>{hero.installCommand}</code>
        </pre>
      </div>
      <p className="hero__status">{hero.status}</p>
    </section>
  );
}
