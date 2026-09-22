import { Footer } from './sections/Footer';
import { Header } from './sections/Header';
import { Hero } from './sections/Hero';
import { HowItWorks } from './sections/HowItWorks';
import { Install } from './sections/Install';
import { Limits } from './sections/Limits';
import { Status } from './sections/Status';

export default function App() {
  return (
    <main>
      <Header />
      <Hero />
      <HowItWorks />
      <Limits />
      <Install />
      <Status />
      <Footer />
    </main>
  );
}
