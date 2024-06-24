import type { Metadata } from "next";
import { Inter } from "next/font/google";
import "./globals.css";
import Navbar from "./navbar";
import Provider from "./provider";

const inter = Inter({ subsets: ["latin"] });

export const metadata: Metadata = {
  title: "LMSYSMD",
  description: "LMSYSMD",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html className={inter.className} lang="en" suppressHydrationWarning>
      <body className="font-sans antialiased bg-background touch-manipulation">
        <Provider>
          <div className="flex relative flex-col min-h-dvh">
            <Navbar />
            <main className="container flex-grow py-12 px-4 mx-auto max-w-6xl">
              {children}
            </main>
          </div>
        </Provider>
      </body>
    </html>
  );
}
