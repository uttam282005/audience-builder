import React, { useState } from 'react';
import {
  EventType,
  Operator,
  FormCondition,
  PreviewRequest,
  EVENT_TYPE_LABELS,
  OPERATOR_LABELS,
} from '../types/audience';
import { Plus, Trash2, Sparkles, Send, RotateCcw } from 'lucide-react';

interface AudienceFormProps {
  onSubmit: (request: PreviewRequest) => void;
  isLoading: boolean;
}

const DEFAULT_SAMPLE_AS_OF = '2026-09-29T00:00:00.000Z';

export const AudienceForm: React.FC<AudienceFormProps> = ({
  onSubmit,
  isLoading,
}) => {
  const [name, setName] = useState<string>('Viewed but not purchased');
  const [asOf, setAsOf] = useState<string>(DEFAULT_SAMPLE_AS_OF);
  const [conditions, setConditions] = useState<FormCondition[]>([
    {
      id: 'cond_1',
      eventType: 'product_view',
      operator: 'at_least',
      count: 2,
      withinDays: 7,
    },
    {
      id: 'cond_2',
      eventType: 'purchase',
      operator: 'exactly',
      count: 0,
      withinDays: 7,
    },
  ]);

  const [formErrors, setFormErrors] = useState<Record<string, string>>({});

  const handleAddCondition = () => {
    const newCond: FormCondition = {
      id: `cond_${Date.now()}`,
      eventType: 'product_view',
      operator: 'at_least',
      count: 1,
      withinDays: 7,
    };
    setConditions([...conditions, newCond]);
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

  const loadSamplePreset = () => {
    setName('Viewed but not purchased');
    setAsOf(DEFAULT_SAMPLE_AS_OF);
    setConditions([
      {
        id: 'cond_1',
        eventType: 'product_view',
        operator: 'at_least',
        count: 2,
        withinDays: 7,
      },
      {
        id: 'cond_2',
        eventType: 'purchase',
        operator: 'exactly',
        count: 0,
        withinDays: 7,
      },
    ]);
    setFormErrors({});
  };

  const loadCurrentTime = () => {
    setAsOf(new Date().toISOString());
  };

  const validate = (): boolean => {
    const errors: Record<string, string> = {};

    if (!name.trim()) {
      errors.name = 'Rule name is required';
    }

    if (!asOf.trim()) {
      errors.asOf = 'Reference timestamp is required';
    } else {
      const parsed = Date.parse(asOf);
      if (isNaN(parsed)) {
        errors.asOf = 'Must be a valid ISO 8601 timestamp (e.g. 2026-09-29T00:00:00.000Z)';
      }
    }

    if (conditions.length === 0) {
      errors.conditions = 'At least one condition is required';
    }

    conditions.forEach((c, idx) => {
      if (c.count < 0 || isNaN(c.count)) {
        errors[`cond_${c.id}_count`] = `Condition #${idx + 1}: count must be ≥ 0`;
      }
      if (c.withinDays < 1 || isNaN(c.withinDays)) {
        errors[`cond_${c.id}_withinDays`] = `Condition #${idx + 1}: lookback must be at least 1 day`;
      }
    });

    setFormErrors(errors);
    return Object.keys(errors).length === 0;
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!validate()) return;

    const requestPayload: PreviewRequest = {
      name: name.trim(),
      asOf: asOf.trim(),
      conditions: conditions.map(({ eventType, operator, count, withinDays }) => ({
        eventType,
        operator,
        count: Number(count),
        withinDays: Number(withinDays),
      })),
    };

    onSubmit(requestPayload);
  };

  return (
    <section className="card" aria-labelledby="rule-builder-heading">
      <div className="card-header">
        <div>
          <h2 id="rule-builder-heading" className="card-title">Audience Definition</h2>
          <p className="card-desc">
            Define logical conditions to segment anonymous visitors. All conditions combine with AND.
          </p>
        </div>
      </div>

      {/* Quick Presets Bar */}
      <div className="presets-bar" role="region" aria-label="Presets">
        <Sparkles size={16} color="#2563eb" aria-hidden="true" />
        <span className="preset-title">Presets:</span>
        <button
          type="button"
          className="preset-button"
          onClick={loadSamplePreset}
        >
          Viewed but not purchased (Sample)
        </button>
        <button
          type="button"
          className="preset-button"
          onClick={() => {
            setName('Purchasers with Cart Activity');
            setAsOf(DEFAULT_SAMPLE_AS_OF);
            setConditions([
              {
                id: 'cond_1',
                eventType: 'purchase',
                operator: 'at_least',
                count: 1,
                withinDays: 7,
              },
            ]);
            setFormErrors({});
          }}
        >
          Converted Buyers
        </button>
      </div>

      <form onSubmit={handleSubmit} noValidate>
        {/* Rule Name */}
        <div className="form-group">
          <label htmlFor="audience-name" className="form-label">
            Audience Name <span aria-hidden="true">*</span>
          </label>
          <input
            id="audience-name"
            type="text"
            className={`form-input ${formErrors.name ? 'error' : ''}`}
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="e.g. Viewed but not purchased"
            aria-required="true"
            aria-invalid={!!formErrors.name}
            aria-describedby={formErrors.name ? 'name-error' : undefined}
          />
          {formErrors.name && (
            <p id="name-error" className="error-text" role="alert">
              {formErrors.name}
            </p>
          )}
        </div>

        {/* Reference asOf */}
        <div className="form-group">
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '6px' }}>
            <label htmlFor="as-of-time" className="form-label" style={{ marginBottom: 0 }}>
              Reference Timestamp (asOf) <span aria-hidden="true">*</span>
            </label>
            <div style={{ display: 'flex', gap: '6px' }}>
              <button
                type="button"
                className="preset-button"
                style={{ padding: '2px 8px', fontSize: '11px' }}
                onClick={() => setAsOf(DEFAULT_SAMPLE_AS_OF)}
              >
                Sample Date (2026-09-29)
              </button>
              <button
                type="button"
                className="preset-button"
                style={{ padding: '2px 8px', fontSize: '11px' }}
                onClick={loadCurrentTime}
              >
                Now
              </button>
            </div>
          </div>
          <input
            id="as-of-time"
            type="text"
            className={`form-input ${formErrors.asOf ? 'error' : ''}`}
            value={asOf}
            onChange={(e) => setAsOf(e.target.value)}
            placeholder="2026-09-29T00:00:00.000Z"
            aria-required="true"
            aria-invalid={!!formErrors.asOf}
            aria-describedby="as-of-hint"
          />
          <p id="as-of-hint" className="hint-text">
            Evaluation is reproducible relative to this point. Events occurring after this time are excluded.
          </p>
          {formErrors.asOf && (
            <p className="error-text" role="alert">
              {formErrors.asOf}
            </p>
          )}
        </div>

        {/* Conditions Section */}
        <div className="conditions-section-title">
          <h3>Conditions ({conditions.length})</h3>
          <button
            type="button"
            className="btn btn-secondary"
            onClick={handleAddCondition}
            style={{ padding: '6px 12px', fontSize: '12px' }}
          >
            <Plus size={14} aria-hidden="true" />
            Add Condition
          </button>
        </div>

        <div className="conditions-list" role="list" aria-label="Rule Conditions">
          {conditions.map((cond, index) => (
            <div key={cond.id} className="condition-row" role="listitem">
              <div className="condition-header">
                <span className="condition-badge">Condition #{index + 1}</span>
                <button
                  type="button"
                  className="btn-danger-ghost"
                  onClick={() => handleRemoveCondition(cond.id)}
                  disabled={conditions.length <= 1}
                  aria-label={`Remove condition ${index + 1}`}
                  title={
                    conditions.length <= 1
                      ? 'At least one condition is required'
                      : 'Remove condition'
                  }
                >
                  <Trash2 size={14} aria-hidden="true" />
                  Remove
                </button>
              </div>

              <div className="condition-grid">
                {/* Event Type */}
                <div className="condition-field">
                  <label htmlFor={`cond-${cond.id}-event`}>Event Type</label>
                  <select
                    id={`cond-${cond.id}-event`}
                    className="form-select"
                    value={cond.eventType}
                    onChange={(e) =>
                      handleConditionChange(
                        cond.id,
                        'eventType',
                        e.target.value as EventType
                      )
                    }
                  >
                    {(Object.keys(EVENT_TYPE_LABELS) as EventType[]).map((type) => (
                      <option key={type} value={type}>
                        {EVENT_TYPE_LABELS[type]}
                      </option>
                    ))}
                  </select>
                </div>

                {/* Operator */}
                <div className="condition-field">
                  <label htmlFor={`cond-${cond.id}-op`}>Operator</label>
                  <select
                    id={`cond-${cond.id}-op`}
                    className="form-select"
                    value={cond.operator}
                    onChange={(e) =>
                      handleConditionChange(
                        cond.id,
                        'operator',
                        e.target.value as Operator
                      )
                    }
                  >
                    {(Object.keys(OPERATOR_LABELS) as Operator[]).map((op) => (
                      <option key={op} value={op}>
                        {OPERATOR_LABELS[op]}
                      </option>
                    ))}
                  </select>
                </div>

                {/* Count */}
                <div className="condition-field">
                  <label htmlFor={`cond-${cond.id}-count`}>Count</label>
                  <input
                    id={`cond-${cond.id}-count`}
                    type="number"
                    min="0"
                    className={`form-input ${
                      formErrors[`cond_${cond.id}_count`] ? 'error' : ''
                    }`}
                    value={cond.count}
                    onChange={(e) =>
                      handleConditionChange(
                        cond.id,
                        'count',
                        parseInt(e.target.value, 10) || 0
                      )
                    }
                  />
                </div>

                {/* Within Days */}
                <div className="condition-field">
                  <label htmlFor={`cond-${cond.id}-days`}>Lookback (Days)</label>
                  <input
                    id={`cond-${cond.id}-days`}
                    type="number"
                    min="1"
                    max="365"
                    className={`form-input ${
                      formErrors[`cond_${cond.id}_withinDays`] ? 'error' : ''
                    }`}
                    value={cond.withinDays}
                    onChange={(e) =>
                      handleConditionChange(
                        cond.id,
                        'withinDays',
                        parseInt(e.target.value, 10) || 1
                      )
                    }
                  />
                </div>
              </div>

              {formErrors[`cond_${cond.id}_count`] && (
                <p className="error-text">
                  {formErrors[`cond_${cond.id}_count`]}
                </p>
              )}
              {formErrors[`cond_${cond.id}_withinDays`] && (
                <p className="error-text">
                  {formErrors[`cond_${cond.id}_withinDays`]}
                </p>
              )}
            </div>
          ))}
        </div>

        {/* Action Controls */}
        <div style={{ display: 'flex', gap: '12px', marginTop: '24px' }}>
          <button
            type="submit"
            className="btn btn-primary"
            disabled={isLoading}
            style={{ flex: 1 }}
          >
            {isLoading ? (
              <>Evaluating Audience...</>
            ) : (
              <>
                <Send size={16} aria-hidden="true" />
                Preview Audience
              </>
            )}
          </button>
          <button
            type="button"
            className="btn btn-secondary"
            onClick={loadSamplePreset}
            disabled={isLoading}
            title="Reset form to default sample values"
          >
            <RotateCcw size={15} aria-hidden="true" />
            Reset
          </button>
        </div>
      </form>
    </section>
  );
};
