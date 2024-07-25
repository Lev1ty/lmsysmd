"use client";

import {
  ClerkLoaded,
  SignedIn,
  SignedOut,
  UserButton,
} from "@clerk/clerk-react";
import { Button, Link } from "@nextui-org/react";

export default function User() {
  return (
    <ClerkLoaded>
      <SignedIn>
        <UserButton />
      </SignedIn>
      <SignedOut>
        <Button
          className="font-sans"
          as={Link}
          color="primary"
          href={`/rating?ts=${new Date().getTime()}`}
          radius="full"
        >
          Sign&nbsp;In
        </Button>
      </SignedOut>
    </ClerkLoaded>
  );
}
