import { hero, install } from '../content';

export function Install() {
  return (
    <section id="install" className="install">
      <h2>{install.heading}</h2>
      <div className="glass snippet">
        <pre>
          <code>{hero.installCommand}</code>
        </pre>
      </div>
      {install.items.map((line) => (
        <p key={line}>{line}</p>
      ))}
    </section>
  );
}
