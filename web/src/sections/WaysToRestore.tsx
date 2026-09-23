import { waysToRestore } from '../content';

export function WaysToRestore() {
  return (
    <section id="ways-to-restore" className="ways-to-restore">
      <h2>{waysToRestore.heading}</h2>
      <ul className="ways">
        {waysToRestore.items.map((item) => (
          <li className="way" key={item}>
            {item}
          </li>
        ))}
      </ul>
    </section>
  );
}
