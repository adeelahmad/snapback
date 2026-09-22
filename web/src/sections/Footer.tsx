import { footerLinks } from '../content';

export function Footer() {
  return (
    <footer className="site-footer">
      {footerLinks.map((link) => (
        <a
          key={link.href}
          href={link.href}
          rel={link.href.startsWith('http') ? 'noopener noreferrer' : undefined}
        >
          {link.label}
        </a>
      ))}
      <p className="footer-note">This site uses Google Analytics to count visits.</p>
    </footer>
  );
}
