import { NextResponse } from 'next/server'
import { getServerSession } from 'next-auth'

// 👥 Mock data for development
const mockTenants = [
  {
    id: '1',
    name: 'Acme Corp',
    description: 'Enterprise customer',
    userCount: 25,
    status: 'active',
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString(),
  },
  {
    id: '2',
    name: 'Startup Inc',
    description: 'Startup customer',
    userCount: 8,
    status: 'active',
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString(),
  },
]

// 📝 GET /api/tenants
export async function GET() {
  const session = await getServerSession()

  if (!session) {
    return new NextResponse(
      JSON.stringify({ error: 'Unauthorized' }),
      { status: 401 }
    )
  }

  // TODO: Replace with actual API call
  return NextResponse.json(mockTenants)
}

// ➕ POST /api/tenants
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
    const { name, description } = body

    if (!name) {
      return new NextResponse(
        JSON.stringify({ error: 'Name is required' }),
        { status: 400 }
      )
    }

    // TODO: Replace with actual API call
    const newTenant = {
      id: Math.random().toString(),
      name,
      description,
      userCount: 0,
      status: 'active',
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
    }

    return NextResponse.json(newTenant)
  } catch (error) {
    return new NextResponse(
      JSON.stringify({ error: 'Invalid request' }),
      { status: 400 }
    )
  }
} 