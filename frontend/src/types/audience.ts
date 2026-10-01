export type EventType =
  | 'page_view'
  | 'product_view'
  | 'add_to_cart'
  | 'checkout_started'
  | 'purchase';

export const EVENT_TYPE_LABELS: Record<EventType, string> = {
  page_view: 'Page View',
  product_view: 'Product View',
  add_to_cart: 'Add to Cart',
  checkout_started: 'Checkout Started',
  purchase: 'Purchase',
};

export type Operator = 'at_least' | 'exactly';

export const OPERATOR_LABELS: Record<Operator, string> = {
  at_least: 'at least (≥)',
  exactly: 'exactly (=)',
};

export interface FormCondition {
  id: string;
  eventType: EventType;
  operator: Operator;
  count: number;
  withinDays: number;
}

export interface ConditionPayload {
  eventType: EventType;
  operator: Operator;
  count: number;
  withinDays: number;
}

export interface PreviewRequest {
  name: string;
  asOf: string;
  conditions: ConditionPayload[];
}

export interface EvidenceItem {
  eventType: EventType;
  observedCount: number;
}

export interface Member {
  anonymousId: string;
  evidence: EvidenceItem[];
}

export interface PreviewResponse {
  name: string;
  asOf: string;
  total: number;
  members: Member[];
}

export interface FieldError {
  field: string;
  issue: string;
}

export interface ApiErrorResponse {
  error: string;
  message: string;
  details?: FieldError[];
}

export interface HealthResponse {
  status: string;
  version: string;
  timestamp: string;
}
