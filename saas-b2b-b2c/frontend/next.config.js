/** @type {import('next').NextConfig} */
const nextConfig = {
  reactStrictMode: false,
  poweredByHeader: false,
  output: 'standalone',
  images: {
    unoptimized: true,
  },
  // Убираем rewrites, чтобы ходить напрямую в API через baseURL axios
};

module.exports = nextConfig;
