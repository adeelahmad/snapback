import { header } from '../content';

export function Header() {
  return (
    <header className="site-header">
      <a href="/" className="site-header__home">
        <picture>
          <source srcSet={header.lockupDark} media="(prefers-color-scheme: dark)" />
          <img src={header.lockupLight} alt={header.lockupAlt} height={32} />
        </picture>
      </a>
      <nav className="site-header__nav">
        {header.links.map((link) => (
          <a key={link.href} href={link.href}>
            {link.label}
          </a>
        ))}
      </nav>
    </header>
  );
}
