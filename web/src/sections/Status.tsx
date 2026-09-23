import { releases, statusHeading } from '../content';

export function Status() {
  return (
    <section id="status" className="status">
      <h2>{statusHeading}</h2>
      <ul className="releases">
        {releases.map(({ name, state, scope }) => (
          <li className="release" key={name}>
            <span className="release__name">{name}</span>{' '}
            <span className="release__state">{state}</span>
            <p className="release__scope">{scope}</p>
          </li>
        ))}
      </ul>
    </section>
  );
}
