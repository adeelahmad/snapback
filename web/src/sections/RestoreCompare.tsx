import { restoreCompare } from '../content';

export function RestoreCompare() {
  const { heading, intro, before, after } = restoreCompare;

  return (
    <section id="restore-compare" className="restore-compare">
      <h2>{heading}</h2>
      <p>{intro}</p>
      <div className="compare">
        <figure className="glass terminal">
          <figcaption className="terminal__label">{before.label}</figcaption>
          <pre>
            {before.lines.map((l, i) => (
              <span key={i} className={l.kind === 'cmd' ? 'terminal__cmd' : 'terminal__out'}>
                {l.text}
                {'\n'}
              </span>
            ))}
          </pre>
        </figure>
        <figure className="glass terminal">
          <figcaption className="terminal__label">{after.label}</figcaption>
          <pre>
            {after.lines.map((l, i) => (
              <span key={i} className={l.kind === 'cmd' ? 'terminal__cmd' : 'terminal__out'}>
                {l.text}
                {'\n'}
              </span>
            ))}
          </pre>
        </figure>
      </div>
    </section>
  );
}
