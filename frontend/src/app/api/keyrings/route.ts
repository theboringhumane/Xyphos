import { NextResponse } from 'next/server'
import { getServerSession } from 'next-auth'

// 🔑 Mock data for development
const mockKeyrings = [
  {
    id: '1',
    name: 'production-keys',
    description: 'Production environment keys',
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
  },
  {
    id: '2',
    name: 'staging-keys',
    description: 'Staging environment keys',
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
  },
]

// 📝 GET /api/keyrings
export async function GET() {
  const session = await getServerSession()

  if (!session) {
    return new NextResponse(
      JSON.stringify({ error: 'Unauthorized' }),
      { status: 401 }
    )
  }

  // TODO: Replace with actual API call
  return NextResponse.json(mockKeyrings)
}

// ➕ POST /api/keyrings
export async function POST(request: Request) {
  const session = await getServerSession()

  if (!session) {
    return new NextResponse(
      JSON.stringify({ error: 'Unauthorized' }),
      { status: 401 }
    )
  }

  try {
    const body = await request.json()
    const { name } = body

    if (!name) {
      return new NextResponse(
        JSON.stringify({ error: 'Name is required' }),
        { status: 400 }
      )
    }

    // TODO: Replace with actual API call
    const newKeyring = {
      id: Math.random().toString(),
      name,
      description: '',
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
    }

    return NextResponse.json(newKeyring)
  } catch (error) {
    return new NextResponse(
      JSON.stringify({ error: 'Invalid request' }),
      { status: 400 }
    )
  }
} 