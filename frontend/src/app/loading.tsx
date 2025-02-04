import { Loader2 } from 'lucide-react'

// ⌛ Loading Page
export default function Loading() {
  return (
    <div className="min-h-screen flex items-center justify-center">
      <div className="flex flex-col items-center space-y-4">
        <Loader2 className="h-12 w-12 animate-spin text-primary-600" />
        <p className="text-lg font-medium text-gray-600">Loading...</p>
      </div>
    </div>
  )
}