import React, { useState } from 'react';
import { PreviewResponse, EVENT_TYPE_LABELS } from '../types/audience';
import { Users, Search, CheckCircle, HelpCircle, Layers } from 'lucide-react';

interface AudienceResultsProps {
  result: PreviewResponse | null;
  isLoading: boolean;
}

export const AudienceResults: React.FC<AudienceResultsProps> = ({
  result,
  isLoading,
}) => {
  const [filterText, setFilterText] = useState<string>('');

  if (isLoading) {
    return (
      <section
        className="results-card"
        aria-labelledby="results-heading"
        aria-busy="true"
        aria-live="polite"
      >
        <div className="card-header">
          <div>
            <h2 id="results-heading" className="card-title">Audience Preview</h2>
            <p className="card-desc">Querying SQLite database...</p>
          </div>
        </div>
        <div className="state-container">
          <div className="spinner" aria-hidden="true" />
          <h3 className="state-title">Evaluating Conditions</h3>
          <p className="state-desc">
            Executing dynamic CTE aggregation across anonymous event history.
          </p>
        </div>
      </section>
    );
  }

  if (!result) {
    return (
      <section className="results-card" aria-labelledby="results-heading">
        <div className="card-header">
          <div>
            <h2 id="results-heading" className="card-title">Audience Preview</h2>
            <p className="card-desc">Ready to evaluate</p>
          </div>
        </div>
        <div className="state-container">
          <Layers size={48} className="state-icon" aria-hidden="true" />
          <h3 className="state-title">No Audience Evaluated Yet</h3>
          <p className="state-desc">
            Configure your rule conditions on the left and click &quot;Preview Audience&quot; to inspect qualified anonymous members.
          </p>
        </div>
      </section>
    );
  }

  const filteredMembers = result.members.filter((m) =>
    m.anonymousId.toLowerCase().includes(filterText.toLowerCase())
  );

  return (
    <section
      className="results-card"
      aria-labelledby="results-heading"
      aria-live="polite"
    >
      <div className="card-header">
        <div>
          <h2 id="results-heading" className="card-title">Audience Preview</h2>
          <p className="card-desc">
            Rule: <strong>{result.name}</strong>
          </p>
        </div>
      </div>

      {/* Summary Metric Banner */}
      <div className="audience-summary-banner">
        <div>
          <div className="summary-metric">
            <span className="metric-number">{result.total}</span>
            <span className="metric-label">
              {result.total === 1 ? 'Matched User' : 'Matched Users'}
            </span>
          </div>
          <div className="metric-timestamp">
            Evaluated relative to: <code>{result.asOf}</code>
          </div>
        </div>
        <Users size={32} color="#2563eb" aria-hidden="true" />
      </div>

      {/* Zero match empty state */}
      {result.total === 0 ? (
        <div className="state-container" style={{ padding: '36px 16px' }}>
          <HelpCircle size={40} className="state-icon" aria-hidden="true" />
          <h3 className="state-title">Zero Matches</h3>
          <p className="state-desc">
            No anonymous users met all specified criteria within their respective lookback windows. Try adjusting your count thresholds or widening the lookback window.
          </p>
        </div>
      ) : (
        <>
          {/* Member Search / Filter Bar if > 3 members */}
          {result.members.length > 3 && (
            <div style={{ marginBottom: '14px', position: 'relative' }}>
              <input
                type="text"
                className="form-input"
                placeholder="Filter by anonymous ID..."
                value={filterText}
                onChange={(e) => setFilterText(e.target.value)}
                style={{ paddingLeft: '32px', fontSize: '13px' }}
                aria-label="Filter matching members by anonymous ID"
              />
              <Search
                size={15}
                color="#94a3b8"
                style={{
                  position: 'absolute',
                  left: '10px',
                  top: '50%',
                  transform: 'translateY(-50%)',
                }}
                aria-hidden="true"
              />
            </div>
          )}

          {/* Members Evidence List */}
          <div className="members-list" role="list" aria-label="Matched Anonymous Members">
            {filteredMembers.map((member) => (
              <div key={member.anonymousId} className="member-item" role="listitem">
                <div className="member-top">
                  <span className="member-id">{member.anonymousId}</span>
                  <span
                    style={{
                      display: 'inline-flex',
                      alignItems: 'center',
                      gap: '4px',
                      fontSize: '11px',
                      color: '#16a34a',
                      fontWeight: 600,
                    }}
                  >
                    <CheckCircle size={13} aria-hidden="true" /> Qualified
                  </span>
                </div>

                <div className="member-evidence" aria-label="Evidence breakdown">
                  {member.evidence.map((ev, idx) => (
                    <span key={idx} className="evidence-pill">
                      <span>{EVENT_TYPE_LABELS[ev.eventType] || ev.eventType}:</span>
                      <span className="evidence-count">{ev.observedCount}</span>
                    </span>
                  ))}
                </div>
              </div>
            ))}

            {filteredMembers.length === 0 && filterText && (
              <p style={{ textAlign: 'center', fontSize: '13px', color: '#64748b', padding: '20px' }}>
                No members match &quot;{filterText}&quot;
              </p>
            )}
          </div>
        </>
      )}
    </section>
  );
};
