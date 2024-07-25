"use client";

import { Spinner } from "@nextui-org/react";
import { useRouter } from "next/navigation";
import { useQueryState } from "nuqs";
import { useEffect } from "react";

export default function Redirect() {
  const router = useRouter();
  const [redirectUrl, _setRedirectUrl] = useQueryState("redirect_url");
  useEffect(() => {
    if (redirectUrl) router.push(redirectUrl);
    else router.push("/");
  }, [redirectUrl, router]);
  return (
    <div className="flex flex-col justify-center items-center w-full h-full">
      <Spinner color="success" size="lg" />
    </div>
  );
}
