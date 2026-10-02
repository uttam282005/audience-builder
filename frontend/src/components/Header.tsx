import React from 'react';

interface HeaderProps {
  asOf: string;
}

const MONTHS = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec'];

export function formatAsOfDate(isoStr: string): string {
  if (!isoStr) return '';
  try {
    const date = new Date(isoStr);
    if (isNaN(date.getTime())) return isoStr;
    return `${date.getUTCDate()} ${MONTHS[date.getUTCMonth()]} ${date.getUTCFullYear()}`;
  } catch {
    return isoStr;
  }
}

export const Header: React.FC<HeaderProps> = ({ asOf }) => {
  const formattedDate = formatAsOfDate(asOf);

  return (
    <header className="app-header" role="banner">
      <div className="app-header-inner">
        <div className="header-brand">
          <span className="brand-title">Mable</span>{' '}
          <span className="brand-subtitle">Audience Builder</span>
        </div>
        <div className="header-asof">
          As of <span className="header-asof-date">{formattedDate || asOf}</span>
        </div>
      </div>
    </header>
  );
};
