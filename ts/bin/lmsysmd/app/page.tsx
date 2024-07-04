import { Icon } from "@iconify/react";
import { Button, Link } from "@nextui-org/react";

export default function Home() {
  return (
    <section className="flex flex-col gap-8 items-center pt-12">
      <h1 className="font-sans text-5xl font-semibold tracking-tighter text-center md:text-7xl md:leading-none md:text-left leading-[0.9]">
        LMSYSMD
      </h1>
      <div className="flex flex-col gap-4 w-full md:flex-row md:justify-center">
        <Button as={Link} color="primary" href="/leaderboard" size="lg">
          Leaderboard
        </Button>
        <Button
          as={Link}
          color="primary"
          endContent={
            <div className="size-6">
              <Icon
                className="stroke-2"
                icon="solar:arrow-right-outline"
                height={24}
                width={24}
              />
            </div>
          }
          href={`/rating?ts=${new Date().getTime()}`}
          size="lg"
          variant="ghost"
        >
          Rating
        </Button>
      </div>
    </section>
  );
}
