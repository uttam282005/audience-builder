import React from 'react';
import { PreviewResponse, ConditionPayload } from '../types/audience';
import { formatAsOfDate } from './Header';

interface AudienceResultsProps {
  result: PreviewResponse | null;
  error: {
    message: string;
    details?: { field: string; issue: string }[];
  } | null;
  lastConditions?: ConditionPayload[];
  onRetry: () => void;
  isLoading: boolean;
}

export const AudienceResults: React.FC<AudienceResultsProps> = ({
  result,
  error,
  lastConditions,
  onRetry,
  isLoading,
}) => {
  return (
    <section
      className={`pane-results ${isLoading ? 'loading' : ''}`}
      aria-labelledby="results-heading"
      aria-live="polite"
    >
      <h2 id="results-heading" className="section-heading">
        Results
      </h2>

      {/* Server / Network Error State */}
      {error && (
        <div className="server-error-banner" role="alert">
          <div className="server-error-message">{error.message}</div>
          <button type="button" className="btn-primary" onClick={onRetry}>
            Retry
          </button>
        </div>
      )}

      {/* Pre-flight resting state (no result yet, no error) */}
      {!error && !result && (
        <div className="results-resting">
          Configure conditions and select Preview audience.
        </div>
      )}

      {/* Result loaded */}
      {!error && result && (
        <>
          <div className="results-header">
            <div className="result-count">{result.total}</div>
            <div className="result-count-label">
              users match this definition as of {formatAsOfDate(result.asOf)}
            </div>
          </div>

          {result.total === 0 ? (
            <div className="results-empty">
              <div className="empty-headline">
                No users match this definition as of {formatAsOfDate(result.asOf)}.
              </div>
              <div className="empty-guidance">
                Try lowering a count or widening a window.
              </div>
            </div>
          ) : (
            <div className="results-table-container">
              <div className="results-meta">
                Showing all {result.total}
              </div>

              <table className="results-table">
                <thead>
                  <tr>
                    <th scope="col" style={{ width: '220px' }}>
                      Anonymous ID
                    </th>
                    <th scope="col">Why they match</th>
                  </tr>
                </thead>
                <tbody>
                  {result.members.map((member, rowIndex) => (
                    <tr
                      key={member.anonymousId}
                      style={
                        {
                          '--row-index': Math.min(rowIndex, 12),
                        } as React.CSSProperties
                      }
                    >
                      <td className="anon-id">{member.anonymousId}</td>
                      <td>
                        <div className="evidence-list">
                          {member.evidence.map((ev, evIdx) => {
                            const matchingCond = lastConditions?.[evIdx];
                            const opLabel =
                              matchingCond?.operator === 'exactly'
                                ? 'exactly'
                                : 'at least';
                            const targetCount = matchingCond?.count ?? 0;
                            const targetDays = matchingCond?.withinDays ?? 7;

                            return (
                              <div key={evIdx} className="evidence-item">
                                {ev.eventType} —{' '}
                                <span className="observed">{ev.observedCount}</span>{' '}
                                observed ({opLabel}{' '}
                                <span className="threshold">{targetCount}</span> in{' '}
                                <span className="threshold">{targetDays}</span> days)
                              </div>
                            );
                          })}
                        </div>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </>
      )}
    </section>
  );
};
