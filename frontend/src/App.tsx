import React, { useState } from 'react';
import { Header } from './components/Header';
import { AudienceForm } from './components/AudienceForm';
import { AudienceResults } from './components/AudienceResults';
import { ErrorAlert } from './components/ErrorAlert';
import { PreviewRequest, PreviewResponse } from './types/audience';
import { previewAudience, ApiError } from './api/client';

export const App: React.FC = () => {
  const [isLoading, setIsLoading] = useState<boolean>(false);
  const [result, setResult] = useState<PreviewResponse | null>(null);
  const [error, setError] = useState<{
    message: string;
    details?: { field: string; issue: string }[];
  } | null>(null);
  const [lastRequest, setLastRequest] = useState<PreviewRequest | null>(null);

  const executePreview = async (request: PreviewRequest) => {
    setIsLoading(true);
    setError(null);
    setLastRequest(request);

    try {
      const data = await previewAudience(request);
      setResult(data);
    } catch (err) {
      if (err instanceof ApiError) {
        setError({
          message: err.message,
          details: err.details,
        });
      } else if (err instanceof Error) {
        setError({ message: err.message });
      } else {
        setError({ message: 'An unexpected error occurred.' });
      }
    } finally {
      setIsLoading(false);
    }
  };

  const handleRetry = () => {
    if (lastRequest) {
      executePreview(lastRequest);
    }
  };

  return (
    <div>
      <Header />

      <main className="container">
        {error && <ErrorAlert error={error} onRetry={handleRetry} />}

        <div className="main-grid">
          <AudienceForm onSubmit={executePreview} isLoading={isLoading} />
          <AudienceResults result={result} isLoading={isLoading} />
        </div>
      </main>
    </div>
  );
};

export default App;
