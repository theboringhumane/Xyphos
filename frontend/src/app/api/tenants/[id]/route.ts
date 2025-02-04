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

// 📝 GET /api/tenants/[id]
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

  const tenant = mockTenants.find((t) => t.id === params.id)

  if (!tenant) {
    return new NextResponse(
      JSON.stringify({ error: 'Tenant not found' }),
      { status: 404 }
    )
  }

  return NextResponse.json(tenant)
}

// 🗑️ DELETE /api/tenants/[id]
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

  const tenant = mockTenants.find((t) => t.id === params.id)

  if (!tenant) {
    return new NextResponse(
      JSON.stringify({ error: 'Tenant not found' }),
      { status: 404 }
    )
  }

  // TODO: Replace with actual API call
  return new NextResponse(null, { status: 204 })
}

// ✏️ PATCH /api/tenants/[id]
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
    const { name, description } = body

    const tenant = mockTenants.find((t) => t.id === params.id)

    if (!tenant) {
      return new NextResponse(
        JSON.stringify({ error: 'Tenant not found' }),
        { status: 404 }
      )
    }

    // TODO: Replace with actual API call
    const updatedTenant = {
      ...tenant,
      name: name || tenant.name,
      description: description || tenant.description,
      updatedAt: new Date().toISOString(),
    }

    return NextResponse.json(updatedTenant)
  } catch (error) {
    return new NextResponse(
      JSON.stringify({ error: 'Invalid request' }),
      { status: 400 }
    )
  }
} 