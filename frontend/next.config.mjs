const rawBackend = process.env.BACKEND_URL || "http://localhost:8080";
const backendUrl = /^https?:\/\//.test(rawBackend) ? rawBackend : `http://${rawBackend}`;

/** @type {import('next').NextConfig} */
const nextConfig = {
  reactStrictMode: true,
  outputFileTracingRoot: process.cwd(),
  async rewrites() {
    return [
      { source: "/api/:path*", destination: `${backendUrl}/api/:path*` },
      { source: "/healthz", destination: `${backendUrl}/healthz` },
    ];
  },
};

export default nextConfig;
