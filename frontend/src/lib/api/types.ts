// 🔑 Keyring type
export interface Keyring {
  id: string;
  name: string;
  description?: string;
  created_at: string;
  updated_at: string;
}

// 🔐 Key type
export interface Key {
  id: string;
  version: string;
  algorithm: string;
  purpose: string;
  status: "active" | "inactive" | "scheduled_for_deletion";
  created_at: string;
  expires_at?: string;
}

// 👥 Tenant type
export interface Tenant {
  id: string;
  name: string;
  description?: string;
  userCount: number;
  status: "active" | "inactive" | "suspended";
  createdAt: string;
  updatedAt: string;
}

// 🎯 API Response types
export interface ApiResponse<T> {
  data: T;
  error?: string;
}

export interface PaginatedResponse<T> {
  data: T[];
  total: number;
  page: number;
  per_page: number;
}

// 🚨 API Error type
export interface ApiError {
  message: string;
  code: string;
  details?: Record<string, any>;
}

// 👥 Tenant type
export interface Tenant {
  id: string;
  name: string;
  description?: string;
  userCount: number;
  status: "active" | "inactive" | "suspended";
  createdAt: string;
  updatedAt: string;
}
