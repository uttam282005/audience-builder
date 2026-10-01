import React from 'react';
import { AlertTriangle, RotateCcw } from 'lucide-react';

interface ErrorAlertProps {
  error: {
    message: string;
    details?: { field: string; issue: string }[];
  };
  onRetry: () => void;
}

export const ErrorAlert: React.FC<ErrorAlertProps> = ({ error, onRetry }) => {
  return (
    <div
      className="alert alert-danger"
      role="alert"
      aria-live="assertive"
      tabIndex={0}
    >
      <AlertTriangle size={20} style={{ flexShrink: 0, marginTop: '2px' }} aria-hidden="true" />
      <div className="alert-content">
        <h4 className="alert-title">Preview Failed</h4>
        <p className="alert-message">{error.message}</p>

        {error.details && error.details.length > 0 && (
          <ul className="alert-details">
            {error.details.map((item, idx) => (
              <li key={idx}>
                <strong>{item.field}:</strong> {item.issue}
              </li>
            ))}
          </ul>
        )}

        <div style={{ marginTop: '12px' }}>
          <button
            type="button"
            className="btn btn-secondary"
            onClick={onRetry}
            style={{
              padding: '6px 14px',
              fontSize: '13px',
              backgroundColor: '#fee2e2',
              borderColor: '#fca5a5',
              color: '#991b1b',
            }}
          >
            <RotateCcw size={14} aria-hidden="true" />
            Retry Request
          </button>
        </div>
      </div>
    </div>
  );
};
