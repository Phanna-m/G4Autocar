export type ApiEnvelope<T> = { data: T };

export type Category = {
  id: string;
  name: string;
  slug: string;
  type: string;
  description?: string;
};

export type Product = {
  id: string;
  category_id: string;
  category_name?: string;
  category_type?: string;
  sku: string;
  name: string;
  brand: string;
  model: string;
  year: number;
  base_price: number;
  condition: string;
  description?: string;
  images?: unknown;
  specs?: unknown;
  available_units?: number;
};

export type Branch = {
  id: string;
  name: string;
  city: string;
  address: string;
  phone: string;
  is_active: boolean;
};

export type User = {
  id: string;
  username: string;
  email: string;
  full_name: string;
  role: string;
  branch_id?: string | null;
};

export type Order = {
  id: string;
  order_no: string;
  customer_name?: string;
  customer_phone?: string;
  branch_name?: string;
  payment_method: string;
  delivery_method: string;
  delivery_address?: string;
  delivery_fee: number;
  delivery_fee_status: string;
  total_amount: number;
  paid_amount: number;
  balance_due: number;
  status: string;
  payment_status: string;
  created_at: string;
  items?: { item_name: string; quantity: number; unit_price: number }[];
};

export type InventoryItem = {
  id: string;
  branch_name?: string;
  product_name?: string;
  quantity: number;
  reserved_qty: number;
  available_qty: number;
  min_threshold: number;
  selling_price: number;
};

type RequestOptions = Omit<RequestInit, "headers"> & {
  token?: string | null;
  headers?: HeadersInit;
};

export async function apiRequest<T>(
  path: string,
  options: RequestOptions = {},
): Promise<T> {
  const { token, headers: requestHeaders, ...init } = options;
  const headers = new Headers(requestHeaders);
  if (init.body && !headers.has("Content-Type")) {
    headers.set("Content-Type", "application/json");
  }
  if (token) headers.set("Authorization", `Bearer ${token}`);

  const response = await fetch(`/api/v1${path}`, {
    ...init,
    headers,
    cache: "no-store",
  });
  const payload: unknown = await response.json().catch(() => null);

  if (!response.ok) {
    const message =
      typeof payload === "object" &&
      payload !== null &&
      "error" in payload &&
      typeof payload.error === "string"
        ? payload.error
        : `Request failed (${response.status})`;
    throw new Error(message);
  }

  return payload as T;
}

export function unwrapData<T>(payload: ApiEnvelope<T>): T {
  return payload.data;
}

export function productImage(images: unknown): string | null {
  let value = images;
  if (typeof value === "string") {
    const raw = value;
    try {
      value = JSON.parse(raw) as unknown;
    } catch {
      return raw.startsWith("http") ? raw : null;
    }
  }
  if (Array.isArray(value)) {
    const image = value.find((item): item is string => typeof item === "string");
    return image ?? null;
  }
  if (typeof value === "object" && value !== null && "url" in value) {
    return typeof value.url === "string" ? value.url : null;
  }
  return null;
}

export function formatMoney(value: number): string {
  return new Intl.NumberFormat("en-US", {
    style: "currency",
    currency: "USD",
    maximumFractionDigits: 0,
  }).format(value);
}
