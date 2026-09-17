import { Text } from "@radix-ui/themes";
import HeroSection from "../components/Hero";
import {
  KBarAnimator,
  KBarPortal,
  KBarPositioner,
  KBarProvider,
  KBarSearch,
  KBarResults,
  useMatches,
  type Action,
} from "kbar";
import { useEffect, useState } from "react";
import { GetAllUsers } from "../helpers/GetUsers"; // Fixed potential typo: GetUses -> GetUsers

function RenderResults() {
  const { results } = useMatches();

  return (
    <KBarResults
      items={results}
      onRender={({ item, active }) =>
        typeof item === "string" ? (
          <div className="px-4 py-2 text-xs uppercase text-gray-400">{item}</div>
        ) : (
          <div
            className={`px-4 py-2 flex items-center justify-between cursor-pointer ${
              active ? "bg-gray-100 dark:bg-gray-800" : "bg-transparent"
            }`}
          >
            {item.name}
          </div>
        )
      }
    />
  );
}

export default function Home() {
    const [actions, setActions] = useState<Action[]>([]);
    
    useEffect(() => {
        GetAllUsers().then((f: Action[]) => {
            setActions(f);
            console.log(f)
        });
    }, []);

    return (
        <KBarProvider actions={actions}>
        <KBarPortal>
            {/* Renders the content outside the root node */}
            <KBarPositioner className="z-50 bg-black/50 backdrop-blur-sm">
            {/* Centers the content */}
            <KBarAnimator className="w-full max-w-xl bg-white dark:bg-gray-900 rounded-xl shadow-2xl overflow-hidden">
                {/* Search input */}
                <KBarSearch className="w-full px-4 py-3 outline-none bg-transparent border-b border-gray-200 dark:border-gray-800" />
                {/* Render search results */}
                <RenderResults />
            </KBarAnimator>
            </KBarPositioner>
        </KBarPortal>

        <div className="w-screen h-screen flex justify-center items-center">
            <HeroSection />
            <Text className="fixed bottom-5 right-5 text-gray-400">
            Muhammad Muhayodin
            </Text>
        </div>
        </KBarProvider>
    );
}