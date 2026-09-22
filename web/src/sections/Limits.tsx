import { limits } from '../content';

export function Limits() {
  return (
    <section id="limits" className="limits">
      <h2>{limits.heading}</h2>
      <ul>
        {limits.items.map((item) => (
          <li key={item}>{item}</li>
        ))}
      </ul>
    </section>
  );
}
