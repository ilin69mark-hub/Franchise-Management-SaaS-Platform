import { NextResponse } from 'next/server';
import type { NextRequest } from 'next/server';

function decodeRole(token: string): string | null {
  try {
    const payload = token.split('.')[1];
    const json = Buffer.from(payload, 'base64').toString();
    const data = JSON.parse(json);
    return data.role || null;
  } catch {
    return null;
  }
}

export function middleware(req: NextRequest) {
  const { pathname } = req.nextUrl;
  const accessToken = req.cookies.get('access_token')?.value || req.cookies.get('__Host-access_token')?.value;

  const isProtected = pathname.startsWith('/admin') || pathname.startsWith('/franchiser') || pathname.startsWith('/dealer');

  if (isProtected && !accessToken) {
    const url = req.nextUrl.clone();
    url.pathname = '/login';
    return NextResponse.redirect(url);
  }

  if (pathname.startsWith('/admin') && accessToken) {
    const role = decodeRole(accessToken);
    if (role !== 'super_admin') {
      const url = req.nextUrl.clone();
      url.pathname = '/login';
      return NextResponse.redirect(url);
    }
  }

  return NextResponse.next();
}

export const config = {
  matcher: ['/admin/:path*', '/franchiser/:path*', '/dealer/:path*', '/salon-manager/:path*'],
};
