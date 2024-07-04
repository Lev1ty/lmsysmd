"use client";

import { ClerkProvider } from "@clerk/clerk-react";
import { dark } from "@clerk/themes";
import { TransportProvider } from "@connectrpc/connect-query";
import { createConnectTransport } from "@connectrpc/connect-web";
import { NextUIProvider } from "@nextui-org/react";
import { createSyncStoragePersister } from "@tanstack/query-sync-storage-persister";
import { QueryCache, QueryClient } from "@tanstack/react-query";
import { ReactQueryDevtools } from "@tanstack/react-query-devtools/build/modern/production.js";
import { PersistQueryClientProvider } from "@tanstack/react-query-persist-client";
import { ThemeProvider, useTheme } from "next-themes";
import { useRouter } from "next/navigation";
import type { ReactNode } from "react";
import { Toaster, toast } from "sonner";

export default function Provider({ children }: { children: ReactNode }) {
  const router = useRouter();
  return (
    <TransportProvider transport={connectTransport}>
      <PersistQueryClientProvider
        client={queryClient}
        persistOptions={{ persister }}
      >
        <NextUIProvider navigate={router.push}>
          <ThemeProvider attribute="class">
            <ProviderDependent>{children}</ProviderDependent>
          </ThemeProvider>
        </NextUIProvider>
      </PersistQueryClientProvider>
    </TransportProvider>
  );
}

function ProviderDependent({ children }: { children: ReactNode }) {
  const { resolvedTheme, theme } = useTheme();
  return (
    <ClerkProvider
      appearance={{
        baseTheme: resolvedTheme === "dark" ? dark : undefined,
        elements: { userButtonAvatarBox: "w-[40px] h-[40px]" },
      }}
      publishableKey={process.env.NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY}
    >
      {children}
      <Toaster
        expand={false}
        position="top-right"
        richColors
        theme={(theme ?? "system") as "light" | "dark" | "system"}
      />
      {process.env.NEXT_PUBLIC_REACT_QUERY_DEVTOOLS === "true" && (
        <ReactQueryDevtools />
      )}
    </ClerkProvider>
  );
}

const queryClient = new QueryClient();

const persister = createSyncStoragePersister({
  storage: typeof window !== "undefined" ? window.localStorage : undefined,
});

const connectTransport = createConnectTransport({
  baseUrl: "/",
  jsonOptions: { ignoreUnknownFields: true },
  useBinaryFormat: true,
  useHttpGet: true,
});
