import { withAuth } from 'next-auth/middleware';
import { NextResponse } from 'next/server';

export default withAuth(
  function middleware(req) {
    // Add custom headers for API requests
    if (req.nextUrl.pathname.startsWith('/api/')) {
      const requestHeaders = new Headers(req.headers);
      requestHeaders.set('x-path-base', req.nextUrl.pathname);
      return NextResponse.next({
        request: {
          headers: requestHeaders,
        },
      });
    }
    return NextResponse.next();
  },
  {
    callbacks: {
      authorized: ({ token }) => !!token,
    },
    pages: {
      signIn: '/auth/signin',
      error: '/auth/error',
    },
  }
);

// Protect all routes except auth and public ones
export const config = {
  matcher: [
    '/((?!auth|_next/static|_next/image|favicon.ico|public|api/auth).*)',
  ],
}; 