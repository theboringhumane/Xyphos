'use client'

import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { useMutation } from '@tanstack/react-query'
import { useRouter } from 'next/navigation'
import { Button } from '@/components/ui/Button'

// 🔐 Form validation schema
const keySchema = z.object({
  name: z.string().min(3, 'Name must be at least 3 characters'),
  description: z.string().min(10, 'Description must be at least 10 characters'),
  algorithm: z.enum(['AES-256-GCM', 'RSA-4096', 'ECDSA-P256']),
  purpose: z.enum(['ENCRYPT_DECRYPT', 'SIGN_VERIFY']),
})

type KeyFormData = z.infer<typeof keySchema>

interface KeyFormProps {
  keyringId: string
  onSuccess?: () => void
}

export function KeyForm({ keyringId, onSuccess }: KeyFormProps) {
  const router = useRouter()
  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<KeyFormData>({
    resolver: zodResolver(keySchema),
    defaultValues: {
      algorithm: 'AES-256-GCM',
      purpose: 'ENCRYPT_DECRYPT',
    },
  })

  // 🔄 Replace with actual API call
  const { mutate: createKey, isLoading } = useMutation({
    mutationFn: async (data: KeyFormData) => {
      // Simulate API call
      await new Promise((resolve) => setTimeout(resolve, 1000))
      console.log('Creating key:', { keyringId, ...data })
    },
    onSuccess: () => {
      onSuccess?.()
      router.push(`/keyrings/${keyringId}`)
    },
  })

  return (
    <form onSubmit={handleSubmit((data) => createKey(data))} className="space-y-6">
      <div>
        <label htmlFor="name" className="block text-sm font-medium text-text-primary">
          Name
        </label>
        <div className="mt-1">
          <input
            type="text"
            id="name"
            {...register('name')}
            className="block w-full rounded-md border-border-light shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm"
          />
          {errors.name && (
            <p className="mt-1 text-sm text-error-DEFAULT">{errors.name.message}</p>
          )}
        </div>
      </div>

      <div>
        <label htmlFor="description" className="block text-sm font-medium text-text-primary">
          Description
        </label>
        <div className="mt-1">
          <textarea
            id="description"
            rows={3}
            {...register('description')}
            className="block w-full rounded-md border-border-light shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm"
          />
          {errors.description && (
            <p className="mt-1 text-sm text-error-DEFAULT">{errors.description.message}</p>
          )}
        </div>
      </div>

      <div>
        <label htmlFor="algorithm" className="block text-sm font-medium text-text-primary">
          Algorithm
        </label>
        <div className="mt-1">
          <select
            id="algorithm"
            {...register('algorithm')}
            className="block w-full rounded-md border-border-light shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm"
          >
            <option value="AES-256-GCM">AES-256-GCM</option>
            <option value="RSA-4096">RSA-4096</option>
            <option value="ECDSA-P256">ECDSA-P256</option>
          </select>
          {errors.algorithm && (
            <p className="mt-1 text-sm text-error-DEFAULT">{errors.algorithm.message}</p>
          )}
        </div>
      </div>

      <div>
        <label htmlFor="purpose" className="block text-sm font-medium text-text-primary">
          Purpose
        </label>
        <div className="mt-1">
          <select
            id="purpose"
            {...register('purpose')}
            className="block w-full rounded-md border-border-light shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm"
          >
            <option value="ENCRYPT_DECRYPT">Encrypt/Decrypt</option>
            <option value="SIGN_VERIFY">Sign/Verify</option>
          </select>
          {errors.purpose && (
            <p className="mt-1 text-sm text-error-DEFAULT">{errors.purpose.message}</p>
          )}
        </div>
      </div>

      <div className="flex justify-end">
        <Button type="submit" isLoading={isLoading}>
          {isLoading ? 'Creating...' : 'Create Key'}
        </Button>
      </div>
    </form>
  )
} 