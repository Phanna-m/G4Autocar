import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "G4 Auto Car | Find your next ride",
  description:
    "Discover your next car with G4 Auto Car and our trusted local branches.",
};

export default function RootLayout({
  children,
}: Readonly<{ children: React.ReactNode }>) {
  return (
    <html lang="en">
      <body>{children}</body>
    </html>
  );
}
