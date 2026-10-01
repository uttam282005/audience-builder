import React, { useEffect, useState } from 'react';
import { checkHealth, API_BASE_URL } from '../api/client';
import { Database } from 'lucide-react';

export const Header: React.FC = () => {
  const [isOnline, setIsOnline] = useState<boolean | null>(null);
  const [checking, setChecking] = useState<boolean>(true);

  const testConnection = async () => {
    setChecking(true);
    try {
      await checkHealth();
      setIsOnline(true);
    } catch {
      setIsOnline(false);
    } finally {
      setChecking(false);
    }
  };

  useEffect(() => {
    testConnection();
    const interval = setInterval(testConnection, 30000); // Check every 30s
    return () => clearInterval(interval);
  }, []);

  return (
    <header className="header" role="banner">
      <div className="header-inner">
        <div className="logo-area">
          <div className="brand-badge" aria-hidden="true">Mable</div>
          <div>
            <h1 className="app-title">Audience Builder</h1>
            <p className="app-subtitle">Behavioral cohort segmentation engine</p>
          </div>
        </div>

        <div className="header-status">
          <Database size={15} className="text-muted" aria-hidden="true" />
          <span>API: <code>{API_BASE_URL}</code></span>
          <span
            className={`status-dot ${
              checking
                ? ''
                : isOnline
                ? 'online'
                : 'offline'
            }`}
            aria-hidden="true"
          />
          <span style={{ fontSize: '12px', fontWeight: 600 }}>
            {checking
              ? 'Checking...'
              : isOnline
              ? 'Connected'
              : 'Backend Unreachable'}
          </span>
          <button
            onClick={testConnection}
            className="btn btn-secondary"
            style={{ padding: '2px 8px', fontSize: '11px' }}
            title="Check backend health"
            aria-label="Refresh backend connection"
          >
            Check
          </button>
        </div>
      </div>
    </header>
  );
};
