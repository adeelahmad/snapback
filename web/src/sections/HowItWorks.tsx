import { howItWorks } from '../content';

export function HowItWorks() {
  return (
    <section className="how">
      <h2>{howItWorks.heading}</h2>
      <p className="how__label">{howItWorks.label}</p>
      <div className="glass terminal">
        <pre>
          {howItWorks.terminalLines.map((line, i) => (
            <span key={line} className={i % 2 === 0 ? 'terminal__cmd' : 'terminal__out'}>
              {line}
              {'\n'}
            </span>
          ))}
        </pre>
      </div>
    </section>
  );
}
