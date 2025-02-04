import NextAuth from 'next-auth'
import GithubProvider from 'next-auth/providers/github'

const API_URL = process.env.NEXT_PUBLIC_API_URL

// 🔐 NextAuth configuration
const handler = NextAuth({
  providers: [
    GithubProvider({
      clientId: process.env.GITHUB_CLIENT_ID!,
      clientSecret: process.env.GITHUB_CLIENT_SECRET!,
      token: {
        url: `${API_URL}/auth/github/callback`,
        async request({ params }) {
          const { code, state } = params
          // 🔑 Exchange code for token with backend
          const response = await fetch(`${API_URL}/auth/github/callback?code=${code}&state=${state}`)
          const data = await response.json()
          return { tokens: { access_token: data.token } }
        }
      },
      userinfo: {
        url: `${API_URL}/auth/github/user`,
        async request({ tokens }) {
          const response = await fetch(`${API_URL}/auth/github/user`, {
            headers: {
              Authorization: `Bearer ${tokens.access_token}`,
            },
          })
          return await response.json()
        },
      },
    }),
  ],
  // 🔒 Custom session configuration 
  session: {
    strategy: 'jwt',
    maxAge: 30 * 24 * 60 * 60, // 30 days
  },
  // ✨ Custom callbacks
  callbacks: {
    async jwt({ token, account }) {
      // Add backend JWT token to the NextAuth token
      if (account) {
        token.accessToken = account.access_token
      }
      return token
    },
    async session({ session, token }) {
      // Add backend JWT token to the session
      return {
        ...session,
        accessToken: token.accessToken,
      }
    },
  },
  // 🎨 Custom pages
  pages: {
    signIn: '/auth/signin', 
    error: '/auth/error',
  },
})

export { handler as GET, handler as POST }