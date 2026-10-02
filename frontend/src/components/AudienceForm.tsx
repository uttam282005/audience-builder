import React, { useState } from 'react';
import {
  EventType,
  Operator,
  FormCondition,
  PreviewRequest,
} from '../types/audience';

interface AudienceFormProps {
  asOf: string;
  onAsOfChange: (asOf: string) => void;
  onSubmit: (request: PreviewRequest) => void;
  isLoading: boolean;
}

const EVENT_OPTIONS: EventType[] = [
  'product_view',
  'purchase',
  'page_view',
  'add_to_cart',
  'checkout_started',
];

const OPERATOR_OPTIONS: { value: Operator; label: string }[] = [
  { value: 'at_least', label: 'at least' },
  { value: 'exactly', label: 'exactly' },
];

export const AudienceForm: React.FC<AudienceFormProps> = ({
  asOf,
  onAsOfChange,
  onSubmit,
  isLoading,
}) => {
  const [name, setName] = useState<string>('Viewed but not purchased');
  const [conditions, setConditions] = useState<FormCondition[]>([
    {
      id: 'c_1',
      eventType: 'product_view',
      operator: 'at_least',
      count: 2,
      withinDays: 7,
    },
    {
      id: 'c_2',
      eventType: 'purchase',
      operator: 'exactly',
      count: 0,
      withinDays: 7,
    },
  ]);

  const [errors, setErrors] = useState<Record<string, string>>({});

  const handleAddCondition = () => {
    const nextId = `c_${Date.now()}`;
    setConditions([
      ...conditions,
      {
        id: nextId,
        eventType: 'product_view',
        operator: 'at_least',
        count: 1,
        withinDays: 7,
      },
    ]);
  };

  const handleRemoveCondition = (id: string) => {
    if (conditions.length <= 1) return;
    setConditions(conditions.filter((c) => c.id !== id));
  };

  const handleConditionChange = <K extends keyof FormCondition>(
    id: string,
    field: K,
    value: FormCondition[K]
  ) => {
    setConditions(
      conditions.map((c) => (c.id === id ? { ...c, [field]: value } : c))
    );
  };

  const validate = (): boolean => {
    const newErrors: Record<string, string> = {};

    if (!name.trim()) {
      newErrors['name'] = 'Audience name is required.';
    }

    if (!asOf.trim()) {
      newErrors['asOf'] = 'As of timestamp is required.';
    } else {
      const parsed = Date.parse(asOf);
      if (isNaN(parsed)) {
        newErrors['asOf'] = 'Expected ISO 8601 date (e.g. 2026-09-29T00:00:00.000Z).';
      }
    }

    if (conditions.length === 0) {
      newErrors['conditions'] = 'At least one condition is required.';
    }

    conditions.forEach((cond, idx) => {
      if (isNaN(cond.count) || cond.count < 0) {
        newErrors[`count_${cond.id}`] = `Condition ${idx + 1}: count must be ≥ 0.`;
      }
      if (isNaN(cond.withinDays) || cond.withinDays < 1) {
        newErrors[`window_${cond.id}`] = `Condition ${idx + 1}: window must be at least 1 day.`;
      }
    });

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!validate()) return;

    onSubmit({
      name: name.trim(),
      asOf: asOf.trim(),
      conditions: conditions.map(({ eventType, operator, count, withinDays }) => ({
        eventType,
        operator,
        count: Number(count),
        withinDays: Number(withinDays),
      })),
    });
  };

  const errorKeys = Object.keys(errors);

  return (
    <section aria-labelledby="define-audience-heading">
      <h2 id="define-audience-heading" className="section-heading">
        Define audience
      </h2>

      {errorKeys.length > 0 && (
        <div className="error-summary" role="alert">
          <div className="error-summary-title">Fix the following before previewing:</div>
          <ul className="error-summary-list">
            {errorKeys.map((key) => (
              <li key={key}>
                <a href={`#field-${key}`} onClick={(e) => {
                  e.preventDefault();
                  document.getElementById(`field-${key}`)?.focus();
                }}>
                  {errors[key]}
                </a>
              </li>
            ))}
          </ul>
        </div>
      )}

      <form onSubmit={handleSubmit} noValidate>
        {/* Audience name */}
        <div className="field-group">
          <label htmlFor="field-name" className="field-label">
            Audience name
          </label>
          <input
            id="field-name"
            type="text"
            className={`field-input ${errors.name ? 'has-error' : ''}`}
            value={name}
            onChange={(e) => setName(e.target.value)}
            aria-invalid={!!errors.name}
            aria-describedby={errors.name ? 'err-name' : undefined}
          />
          {errors.name && (
            <div id="err-name" className="field-error-message">
              {errors.name}
            </div>
          )}
        </div>

        {/* As of (UTC) */}
        <div className="field-group">
          <label htmlFor="field-asOf" className="field-label">
            As of (UTC)
          </label>
          <input
            id="field-asOf"
            type="text"
            className={`field-input ${errors.asOf ? 'has-error' : ''}`}
            value={asOf}
            onChange={(e) => onAsOfChange(e.target.value)}
            aria-invalid={!!errors.asOf}
            aria-describedby={errors.asOf ? 'err-asOf' : undefined}
          />
          {errors.asOf && (
            <div id="err-asOf" className="field-error-message">
              {errors.asOf}
            </div>
          )}
        </div>

        {/* Conditions list */}
        <div className="field-group" style={{ marginBottom: 'var(--space-4)' }}>
          {conditions.map((cond, index) => (
            <React.Fragment key={cond.id}>
              {index > 0 && (
                <div className="and-separator" aria-hidden="true">
                  AND
                </div>
              )}

              <fieldset className="condition">
                <legend>Condition {index + 1}</legend>

                <div className="condition-grid">
                  {/* Row 1: Event type / Operator */}
                  <div className="condition-field">
                    <label htmlFor={`field-event-${cond.id}`} className="field-label">
                      Event type
                    </label>
                    <select
                      id={`field-event-${cond.id}`}
                      className="field-select"
                      value={cond.eventType}
                      onChange={(e) =>
                        handleConditionChange(cond.id, 'eventType', e.target.value as EventType)
                      }
                    >
                      {EVENT_OPTIONS.map((opt) => (
                        <option key={opt} value={opt}>
                          {opt}
                        </option>
                      ))}
                    </select>
                  </div>

                  <div className="condition-field">
                    <label htmlFor={`field-op-${cond.id}`} className="field-label">
                      Operator
                    </label>
                    <select
                      id={`field-op-${cond.id}`}
                      className="field-select"
                      value={cond.operator}
                      onChange={(e) =>
                        handleConditionChange(cond.id, 'operator', e.target.value as Operator)
                      }
                    >
                      {OPERATOR_OPTIONS.map((opt) => (
                        <option key={opt.value} value={opt.value}>
                          {opt.label}
                        </option>
                      ))}
                    </select>
                  </div>

                  {/* Row 2: Count / Window (days) */}
                  <div className="condition-field">
                    <label htmlFor={`field-count_${cond.id}`} className="field-label">
                      Count
                    </label>
                    <input
                      id={`field-count_${cond.id}`}
                      type="number"
                      min="0"
                      className={`field-input field-input--numeric ${
                        errors[`count_${cond.id}`] ? 'has-error' : ''
                      }`}
                      value={isNaN(cond.count) ? '' : cond.count}
                      onChange={(e) =>
                        handleConditionChange(
                          cond.id,
                          'count',
                          e.target.value === '' ? (NaN as unknown as number) : parseInt(e.target.value, 10)
                        )
                      }
                      aria-invalid={!!errors[`count_${cond.id}`]}
                    />
                    {errors[`count_${cond.id}`] && (
                      <div className="field-error-message">
                        {errors[`count_${cond.id}`]}
                      </div>
                    )}
                  </div>

                  <div className="condition-field">
                    <label htmlFor={`field-window_${cond.id}`} className="field-label">
                      Window (days)
                    </label>
                    <input
                      id={`field-window_${cond.id}`}
                      type="number"
                      min="1"
                      className={`field-input field-input--numeric ${
                        errors[`window_${cond.id}`] ? 'has-error' : ''
                      }`}
                      value={isNaN(cond.withinDays) ? '' : cond.withinDays}
                      onChange={(e) =>
                        handleConditionChange(
                          cond.id,
                          'withinDays',
                          e.target.value === '' ? (NaN as unknown as number) : parseInt(e.target.value, 10)
                        )
                      }
                      aria-invalid={!!errors[`window_${cond.id}`]}
                    />
                    {errors[`window_${cond.id}`] && (
                      <div className="field-error-message">
                        {errors[`window_${cond.id}`]}
                      </div>
                    )}
                  </div>
                </div>

                <div className="condition-actions">
                  <button
                    type="button"
                    className="btn-remove"
                    onClick={() => handleRemoveCondition(cond.id)}
                    disabled={conditions.length <= 1}
                    aria-label={`Remove condition ${index + 1}`}
                  >
                    Remove
                  </button>
                </div>
              </fieldset>
            </React.Fragment>
          ))}
        </div>

        {/* Add condition button */}
        <div className="field-group" style={{ marginBottom: 'var(--space-5)' }}>
          <button
            type="button"
            className="btn-secondary"
            onClick={handleAddCondition}
          >
            Add condition
          </button>
        </div>

        {/* Preview button */}
        <div>
          <button
            type="submit"
            className="btn-primary"
            disabled={isLoading}
            aria-busy={isLoading}
            style={{ width: '100%' }}
          >
            {isLoading ? 'Calculating…' : 'Preview audience'}
          </button>
        </div>
      </form>
    </section>
  );
};
