import { twoProblems } from '../content';

export function TwoProblems() {
  const { heading, intro, backup, restore, closing } = twoProblems;

  return (
    <section id="two-problems" className="two-problems">
      <h2>{heading}</h2>
      <p>{intro}</p>
      <div className="problems">
        <article className="problem problem--backup">
          <h3>{backup.name}</h3>
          <span className="problem__tag">{backup.tag}</span>
          <p>{backup.body}</p>
        </article>
        <article className="problem problem--restore">
          <h3>{restore.name}</h3>
          <span className="problem__tag">{restore.tag}</span>
          <p>{restore.body}</p>
        </article>
      </div>
      <p className="two-problems__closing">{closing}</p>
    </section>
  );
}
