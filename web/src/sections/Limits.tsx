import { limits, resticKeeps } from '../content';

export function Limits() {
  return (
    <section id="limits" className="limits">
      <h2>{resticKeeps.heading}</h2>
      <div className="limits__lists">
        <div>
          <ul>
            {resticKeeps.items.map((item) => (
              <li key={item}>{item}</li>
            ))}
          </ul>
        </div>
        <div>
          <h3>{limits.heading}</h3>
          <ul>
            {limits.items.map((item) => (
              <li key={item}>{item}</li>
            ))}
          </ul>
        </div>
      </div>
    </section>
  );
}
