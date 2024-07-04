import {
  NavbarContent,
  NavbarItem,
  Navbar as NextUINavbar,
} from "@nextui-org/react";
import Brand from "./brand";
import User from "./user";

export default function Navbar() {
  return (
    <NextUINavbar
      classNames={{
        base: "py-4 backdrop-filter-none bg-transparent",
        wrapper: "w-full justify-center bg-transparent",
      }}
      height="54px"
    >
      <NavbarContent
        className="gap-2 px-2 rounded-full bg-background/60 shadow-medium backdrop-blur-md backdrop-saturate-150 dark:bg-default-100/50"
        justify="center"
      >
        <Brand />
        <NavbarItem className="h-[40px]">
          <User />
        </NavbarItem>
      </NavbarContent>
    </NextUINavbar>
  );
}
