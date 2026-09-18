import { Text } from "@radix-ui/themes";
import HeroSection from "../components/Hero";

/**
 * Home route layout only: the global palette, data loading and theme live in App.
 * Forward the current appearance and toggle callback to the hero rather than
 * keeping a separate preference that could drift from the palette's command.
 * min-h-screen allows the page to grow when cards stack on small screens.
 */
export default function Home({ dark, onToggleTheme }: { dark: boolean; onToggleTheme: () => void }) {
  return (
    <main className="min-h-screen flex justify-center items-center p-6">
      <HeroSection dark={dark} onToggleTheme={onToggleTheme} />
      <Text className="fixed bottom-5 right-5 text-gray-400">Muhammad Muhayodin</Text>
    </main>
  );
}