import {Flex, Section, Text, Card, Grid} from "@radix-ui/themes"
import { useKBar } from "kbar"
import { Moon, Plus, Search, User, type LucideIcon } from "lucide-react"

function CardOption(props: {Icon: LucideIcon, text: string, onClick?: () => void}) {
    return <Card className="hover:bg-gray-200 hover:shadow-lg transition-all duration-200 active:bg-gray-400" onClick={props.onClick}>
        <Flex direction="column" align="center" justify="center" gap="3" p="3">
        <props.Icon size="30" />
        <Text>{props.text}</Text>
        </Flex>
    </Card>
}

export default function HeroSection(props: {
}){
    const {query} = useKBar()

    return <Section className="flex flex-col gap-9">
        <Text size={"9"}>Welcome To Generations!</Text>
        <Grid columns={{ initial: "1", sm: "2", md: "4" }} gap="4" width="100%">
            <CardOption text="Search User" Icon={Search} onClick={() => {query.toggle()}} />
            <CardOption text="List Of Singles" Icon={User} />
            <CardOption text="New Family" Icon={Plus} />
            <CardOption text="Dark Mode" Icon={Moon} />
        </Grid>
    </Section>
}