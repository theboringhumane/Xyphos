import NextAuth from 'next-auth'

declare module 'next-auth' {
  interface Session {
    accessToken?: string
    githubId?: string
    user?: {
      id: string
      email: string
      name: string
      image: string
    }
  }
} 