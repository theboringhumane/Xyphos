'use client'

import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { useRouter } from 'next/navigation'
import { Button } from '@/components/ui/Button'
import { useCreateKeyring } from '@/lib/api/hooks'

// 🔑 Form validation schema
const keyringSchema = z.object({
  name: z.string().min(3, 'Name must be at least 3 characters'),
  description: z.string().optional(),
})

type KeyringFormData = z.infer<typeof keyringSchema>

interface KeyringFormProps {
  initialData?: Partial<KeyringFormData>
  onSuccess?: () => void
}

export function KeyringForm({ initialData, onSuccess }: KeyringFormProps) {
  const router = useRouter()
  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<KeyringFormData>({
    resolver: zodResolver(keyringSchema),
    defaultValues: initialData,
  })

  // 🔄 Use the keyring creation hook
  const { mutate: createKeyring, isPending } = useCreateKeyring()

  const onSubmit = (data: KeyringFormData) => {
    createKeyring(data, {
      onSuccess: () => {
        onSuccess?.()
        router.push('/keyrings')
      },
    })
  }

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
      <div>
        <label htmlFor="name" className="block text-sm font-medium text-gray-900">
          Name
        </label>
        <div className="mt-1">
          <input
            type="text"
            id="name"
            {...register('name')}
            className="block w-full rounded-md border-gray-300 shadow-sm focus:border-primary focus:ring-primary sm:text-sm"
            placeholder="Production Keys"
          />
          {errors.name && (
            <p className="mt-1 text-sm text-destructive">{errors.name.message}</p>
          )}
        </div>
      </div>

      <div>
        <label htmlFor="description" className="block text-sm font-medium text-gray-900">
          Description
        </label>
        <div className="mt-1">
          <textarea
            id="description"
            rows={3}
            {...register('description')}
            className="block w-full rounded-md border-gray-300 shadow-sm focus:border-primary focus:ring-primary sm:text-sm"
            placeholder="Keys used for production environment encryption"
          />
          {errors.description && (
            <p className="mt-1 text-sm text-destructive">{errors.description.message}</p>
          )}
        </div>
      </div>

      <div className="flex justify-end">
        <Button type="submit" disabled={isPending}>
          {isPending ? 'Creating...' : 'Create Keyring'}
        </Button>
      </div>
    </form>
  )
} 