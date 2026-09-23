import { Footer } from './sections/Footer';
import { Header } from './sections/Header';
import { Hero } from './sections/Hero';
import { HowItWorks } from './sections/HowItWorks';
import { Install } from './sections/Install';
import { Limits } from './sections/Limits';
import { RestoreCompare } from './sections/RestoreCompare';
import { Status } from './sections/Status';
import { TwoProblems } from './sections/TwoProblems';
import { WaysToRestore } from './sections/WaysToRestore';

export default function App() {
  return (
    <main>
      <Header />
      <Hero />
      <TwoProblems />
      <RestoreCompare />
      <HowItWorks />
      <WaysToRestore />
      <Limits />
      <Install />
      <Status />
      <Footer />
    </main>
  );
}
