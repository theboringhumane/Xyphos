'use client'

import {
  Toast,
  ToastClose,
  ToastDescription,
  ToastProvider,
  ToastTitle,
  ToastViewport,
} from '@radix-ui/react-toast'
import { cva } from 'class-variance-authority'
import { X } from 'lucide-react'
import { useToast } from '@/hooks/useToast'
import { cn } from '@/utils/cn'

// 🎨 Toast variants styling
const toastVariants = cva(
  'group pointer-events-auto relative flex w-full items-center justify-between space-x-4 overflow-hidden rounded-md border p-6 pr-8 shadow-lg transition-all',
  {
    variants: {
      variant: {
        default: 'bg-background border-border-light',
        success: 'bg-success-50 border-success-200 text-success-900',
        error: 'bg-error-50 border-error-200 text-error-900',
      },
    },
    defaultVariants: {
      variant: 'default',
    },
  }
)

// 🔔 Toaster component
export function Toaster() {
  const { toasts } = useToast()

  return (
    <ToastProvider>
      {toasts.map(function ({ id, title, description, action, ...props }) {
        return (
          <Toast key={id} {...props}>
            <div className="grid gap-1">
              {title && <ToastTitle>{title}</ToastTitle>}
              {description && (
                <ToastDescription>{description}</ToastDescription>
              )}
            </div>
            {action}
            <ToastClose className="absolute right-2 top-2 rounded-md p-1">
              <X className="h-4 w-4" />
            </ToastClose>
          </Toast>
        )
      })}
      <ToastViewport />
    </ToastProvider>
  )
} 