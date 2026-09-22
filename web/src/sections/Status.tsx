import { stages, statusHeading } from '../content';

export function Status() {
  return (
    <section id="status" className="status">
      <h2>{statusHeading}</h2>
      <ol start={0}>
        {stages.map((stage) => (
          <li key={stage.name}>
            {stage.name}: <span className="status__state">{stage.state}</span>
          </li>
        ))}
      </ol>
    </section>
  );
}
