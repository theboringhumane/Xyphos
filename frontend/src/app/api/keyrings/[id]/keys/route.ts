import { NextResponse } from 'next/server'
import { getServerSession } from 'next-auth'

// 🔐 Mock data for development
const mockKeys = {
  'production-keys': [
    {
      id: '1',
      version: 'v1',
      algorithm: 'AES-256-GCM',
      purpose: 'encryption',
      status: 'active',
      created_at: new Date().toISOString(),
      expires_at: new Date(Date.now() + 30 * 24 * 60 * 60 * 1000).toISOString(), // 30 days
    },
    {
      id: '2',
      version: 'v2',
      algorithm: 'AES-256-GCM',
      purpose: 'encryption',
      status: 'inactive',
      created_at: new Date(Date.now() - 30 * 24 * 60 * 60 * 1000).toISOString(), // 30 days ago
      expires_at: new Date().toISOString(),
    },
  ],
}

// 📝 GET /api/keyrings/[id]/keys
export async function GET(
  request: Request,
  { params }: { params: { id: string } }
) {
  const session = await getServerSession()

  if (!session) {
    return new NextResponse(
      JSON.stringify({ error: 'Unauthorized' }),
      { status: 401 }
    )
  }

  const keys = mockKeys[params.id as keyof typeof mockKeys] || []
  return NextResponse.json({ keys })
}

// ➕ POST /api/keyrings/[id]/keys
export async function POST(
  request: Request,
  { params }: { params: { id: string } }
) {
  const session = await getServerSession()

  if (!session) {
    return new NextResponse(
      JSON.stringify({ error: 'Unauthorized' }),
      { status: 401 }
    )
  }

  try {
    const body = await request.json()
    const { algorithm, purpose } = body

    if (!algorithm || !purpose) {
      return new NextResponse(
        JSON.stringify({ error: 'Algorithm and purpose are required' }),
        { status: 400 }
      )
    }

    // TODO: Replace with actual API call
    const newKey = {
      id: Math.random().toString(),
      version: `v${Date.now()}`,
      algorithm,
      purpose,
      status: 'active',
      created_at: new Date().toISOString(),
      expires_at: new Date(Date.now() + 30 * 24 * 60 * 60 * 1000).toISOString(), // 30 days
    }

    return NextResponse.json(newKey)
  } catch (error) {
    return new NextResponse(
      JSON.stringify({ error: 'Invalid request' }),
      { status: 400 }
    )
  }
} 