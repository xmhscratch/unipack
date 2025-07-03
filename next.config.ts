import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  /* config options here */
  allowedDevOrigins: ['192.168.56.55'],
  eslint: {
    // Only run ESLint on the 'pages' and 'utils' directories during production builds (next build)
    dirs: ['app'],
  },
};

export default nextConfig;
