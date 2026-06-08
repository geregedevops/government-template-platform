/** @type {import('next').NextConfig} */

// Content-Security-Policy-г middleware.ts-ээс per-request nonce-тэй өгнө.
// Энд зөвхөн CSP-аас бусад статик security header-уудыг хариуцна.

const nextConfig = {
  reactStrictMode: true,
  // Standalone output → slim production Docker image (server.js + minimal
  // node_modules) instead of shipping the whole tree. See frontend/Dockerfile.
  output: 'standalone',
  // Стандарт JWKS зам (/.well-known/jwks.json)-ийг BFF route руу дотооддоо
  // дахин бичнэ — федерацийн node-ууд Gerege-ийн нийтийн түлхүүрийг эндээс авна.
  async rewrites() {
    return [
      { source: '/.well-known/jwks.json', destination: '/api/jwks' },
      { source: '/.well-known/fed-jwks.json', destination: '/api/fed-jwks' },
    ];
  },
  // Хуучин root-level зам → /admin/* (Admin системийг /admin дор төвлөрүүлсэн)
  // тул хуучин холбоос/bookmark эвдрэхгүй.
  async redirects() {
    return [
      { source: '/chat', destination: '/admin/chat', permanent: false },
      { source: '/knowledge', destination: '/admin/knowledge', permanent: false },
      { source: '/translate', destination: '/admin/translate', permanent: false },
      { source: '/settings', destination: '/admin/settings', permanent: false },
      { source: '/profile', destination: '/admin/profile', permanent: false },
      { source: '/bpm', destination: '/admin/bpm', permanent: false },
      { source: '/bpm/:path*', destination: '/admin/bpm/:path*', permanent: false },
      // Personal систем → User систем; Manager bare → /manager/dashboard.
      { source: '/personal', destination: '/user/dashboard', permanent: false },
      { source: '/personal/profile', destination: '/user/profile', permanent: false },
      { source: '/personal/:path*', destination: '/user/:path*', permanent: false },
      { source: '/manager', destination: '/manager/dashboard', permanent: false },
      { source: '/user', destination: '/user/dashboard', permanent: false },
    ];
  },
  // Security headers applied to every response.
  async headers() {
    const securityHeaders = [
      { key: 'X-Frame-Options', value: 'DENY' },
      { key: 'X-Content-Type-Options', value: 'nosniff' },
      { key: 'Referrer-Policy', value: 'strict-origin-when-cross-origin' },
      {
        // microphone=(self) — дуу хоолойн орчуулга + чатын дуу таних (STT)
        // нь mic ашигладаг тул өөрийн origin-д зөвшөөрнө. camera/geolocation
        // хэрэггүй тул хаалттай хэвээр.
        key: 'Permissions-Policy',
        value: 'camera=(), microphone=(self), geolocation=()',
      },
      // Content-Security-Policy — middleware.ts дотор per-request nonce-тэй
      // тохируулагдана. Энд static header болгож тавихгүй (nonce-гүй болно).
    ];

    // HSTS-ийг зөвхөн production-д илгээнэ — dev дээр http тул HSTS тохиромжгүй.
    if (process.env.NODE_ENV === 'production') {
      securityHeaders.push({
        key: 'Strict-Transport-Security',
        value: 'max-age=63072000; includeSubDomains',
      });
    }

    return [
      {
        source: '/:path*',
        headers: securityHeaders,
      },
    ];
  },
};

export default nextConfig;
