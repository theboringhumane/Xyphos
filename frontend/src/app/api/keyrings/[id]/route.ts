import { NextResponse } from 'next/server'
import { getServerSession } from 'next-auth'

// 🔑 Mock data for development
const mockKeyrings = [
  {
    id: 'production-keys',
    name: 'Production Keys',
    description: 'Production environment keys',
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
  },
  {
    id: 'staging-keys',
    name: 'Staging Keys',
    description: 'Staging environment keys',
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
  },
]

// 📝 GET /api/keyrings/[id]
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

  const keyring = mockKeyrings.find((k) => k.id === params.id)

  if (!keyring) {
    return new NextResponse(
      JSON.stringify({ error: 'Keyring not found' }),
      { status: 404 }
    )
  }

  return NextResponse.json(keyring)
}

// 🗑️ DELETE /api/keyrings/[id]
export async function DELETE(
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

  const keyring = mockKeyrings.find((k) => k.id === params.id)

  if (!keyring) {
    return new NextResponse(
      JSON.stringify({ error: 'Keyring not found' }),
      { status: 404 }
    )
  }

  // TODO: Replace with actual API call
  return new NextResponse(null, { status: 204 })
}

// ✏️ PATCH /api/keyrings/[id]
export async function PATCH(
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
    const { description } = body

    const keyring = mockKeyrings.find((k) => k.id === params.id)

    if (!keyring) {
      return new NextResponse(
        JSON.stringify({ error: 'Keyring not found' }),
        { status: 404 }
      )
    }

    // TODO: Replace with actual API call
    const updatedKeyring = {
      ...keyring,
      description,
      updated_at: new Date().toISOString(),
    }

    return NextResponse.json(updatedKeyring)
  } catch (error) {
    return new NextResponse(
      JSON.stringify({ error: 'Invalid request' }),
      { status: 400 }
    )
  }
} 