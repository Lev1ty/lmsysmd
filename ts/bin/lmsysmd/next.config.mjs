/** @type {import('next').NextConfig} */
const nextConfig = {
  images: { unoptimized: true },
  output: "export",
  trailingSlash: true,
  transpilePackages: ["lib"],
};

export default nextConfig;
