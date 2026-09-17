import {Flex, Section, Text, Card, Grid} from "@radix-ui/themes"
import { Plus, Search, User } from "lucide-react"

export default function HeroSection(){
    return <Section className="flex flex-col gap-9">
        <Text size={"9"}>Welcome To Generations!</Text>
        <Grid columns={{ initial: "1", sm: "2", md: "4" }} gap="4" width="100%">
            <Card>
                <Flex direction="column" align="center" justify="center" gap="3" p="3">
                <Search size="30" />
                <Text>Search User</Text>
                </Flex>
            </Card>

            <Card>
                <Flex direction="column" align="center" justify="center" gap="3" p="3">
                <User size="30" />
                <Text>List Of Singles</Text>
                </Flex>
            </Card>

            <Card>
                <Flex direction="column" align="center" justify="center" gap="3" p="3">
                <Plus size="30" />
                <Text>New Family</Text>
                </Flex>
            </Card>

            <Card>
                <Flex direction="column" align="center" justify="center" gap="3" p="3">
                <Search size="30" />
                <Text>Search User</Text>
                </Flex>
            </Card>
        </Grid>
    </Section>
}