import { Flex, Section, Text, Card, Grid } from "@radix-ui/themes";
import { useKBar } from "kbar";
import { Moon, Network, Plus, Search, Sun, User, type LucideIcon } from "lucide-react";
import { useNavigate } from "react-router";

/**
 * Presents a home action as a native button styled by Radix Card.asChild.
 * Native buttons provide keyboard activation and disabled semantics; cards with
 * no handler are unavailable features, not clickable controls that do nothing.
 * The icon is decorative because the visible label supplies the accessible name.
 */
function CardOption({ Icon, text, onClick }: { Icon: LucideIcon; text: string; onClick?: () => void }) {
  return (
    <Card asChild>
      <button type="button" onClick={onClick} disabled={!onClick} className="cursor-pointer disabled:cursor-not-allowed disabled:opacity-50">
        <Flex direction="column" align="center" justify="center" gap="3" p="3">
          <Icon size={30} aria-hidden="true" />
          <Text>{text}</Text>
        </Flex>
      </button>
    </Card>
  );
}

/**
 * Home-page entry points. Theme state and its toggle come from App so the card
 * stays synchronized with the palette's theme action. Search uses the same Kbar
 * controller as the global keyboard shortcut, not a separate modal instance.
 */
export default function HeroSection({ dark, onToggleTheme }: { dark: boolean; onToggleTheme: () => void }) {
  const { query } = useKBar();
  const navigate = useNavigate();
  return (
    <Section className="flex flex-col gap-9">
      <Text size="9">Welcome To Generations!</Text>
      {/* Expand as space becomes available: two columns on small screens and
          three from md up, so the five cards fill two rows instead of one
          cramped line. Every card here opens something. */}
      <Grid columns={{ initial: "1", sm: "2", md: "3" }} gap="4" width="100%">
        <CardOption text="Search User" Icon={Search} onClick={() => query.toggle()} />
        <CardOption text="List Of Singles" Icon={User} onClick={() => navigate("/singles")} />
        <CardOption text="New Family" Icon={Plus} onClick={() => navigate("/families/new")} />
        <CardOption text="Relationships" Icon={Network} onClick={() => navigate("/relationships")} />
        <CardOption text={dark ? "Light Mode" : "Dark Mode"} Icon={dark ? Sun : Moon} onClick={onToggleTheme} />
      </Grid>
      <Text size="2" color="gray">Press Ctrl+K / ⌘K to search people and commands.</Text>
    </Section>
  );
}