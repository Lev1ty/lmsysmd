import { Link, NavbarBrand } from "@nextui-org/react";

export default function Brand() {
  return (
    <NavbarBrand
      as={Link}
      // @ts-ignore
      color="foreground"
      href="/"
    >
      <p className="mx-2 font-sans font-semibold tracking-tighter">LMSYSMD</p>
    </NavbarBrand>
  );
}
