import { NextResponse } from 'next/server';
import type { NextRequest } from 'next/server';

// RE-AUDIT: роль больше НЕ читаем из неподписанного JWT (atob без проверки
// подписи открывал /admin любому с самописным токеном). Edge не знает
// JWT_SECRET и проверить подпись не может — гейт только по наличию сессии,
// роль проверяет API на каждый запрос.
export function middleware(req: NextRequest) {
  const { pathname } = req.nextUrl;
  const accessToken = req.cookies.get('access_token')?.value || req.cookies.get('__Host-access_token')?.value;

  const isProtected = pathname.startsWith('/admin') || pathname.startsWith('/franchiser') || pathname.startsWith('/dealer');

  if (isProtected && !accessToken) {
    const url = req.nextUrl.clone();
    url.pathname = '/login';
    return NextResponse.redirect(url);
  }

  return NextResponse.next();
}

export const config = {
  matcher: ['/admin/:path*', '/franchiser/:path*', '/dealer/:path*', '/salon-manager/:path*'],
};
