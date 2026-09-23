import { hero } from '../content';

export function Hero() {
  return (
    <section id="hero" className="hero">
      <div className="hero__text">
        <p className="hero__eyebrow">{hero.eyebrow}</p>
        <h1>{hero.pitch}</h1>
        <p className="hero__lede">{hero.lede}</p>
        <div className="glass snippet">
          <pre>
            <code>{hero.installCommand}</code>
          </pre>
        </div>
        <p className="hero__status">{hero.status}</p>
      </div>
      <div className="hero__demo">
        {hero.demo.map((pane, i) => (
          <figure key={i} className="glass terminal">
            <figcaption className="terminal__label">{pane.label}</figcaption>
            <pre>
              {pane.lines.map((l, j) => (
                <span key={j} className={l.kind === 'cmd' ? 'terminal__cmd' : 'terminal__out'}>
                  {l.text}
                  {'\n'}
                </span>
              ))}
            </pre>
          </figure>
        ))}
      </div>
    </section>
  );
}
