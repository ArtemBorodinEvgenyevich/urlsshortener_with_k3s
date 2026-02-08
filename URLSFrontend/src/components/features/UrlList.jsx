import { useState, useEffect } from 'react';
import { getUserUrls, deleteUrl } from '../../services/api.js';
import { useCopyToClipboard } from '../../hooks/useCopyToClipboard.js';
import './UrlList.css';

function truncateUrl(url, maxLength = 256) {
  if (url.length <= maxLength) return url;
  return url.substring(0, maxLength) + '...';
}

function formatTimeRemaining(expiresAt) {
  const now = new Date();
  const expires = new Date(expiresAt);
  const diff = expires - now;

  if (diff <= 0) return 'expired';

  const hours = Math.floor(diff / (1000 * 60 * 60));
  const minutes = Math.floor((diff % (1000 * 60 * 60)) / (1000 * 60));

  if (hours > 24) {
    const days = Math.floor(hours / 24);
    return `${days}d`;
  }
  if (hours > 0) {
    return `${hours}h ${minutes}m`;
  }
  return `${minutes}m`;
}

function UrlItem({ url, onDelete, onCopy, copiedCode }) {
  const shortUrl = `${window.location.origin}/${url.short_code}`;
  const isCopied = copiedCode === url.short_code;
  const timeRemaining = formatTimeRemaining(url.expires_at);
  const isExpired = timeRemaining === 'expired';

  const handleDelete = async () => {
    try {
      await deleteUrl(url.short_code);
      onDelete(url.short_code);
    } catch (error) {
      console.error('Failed to delete:', error);
    }
  };

  return (
    <div className={`url-list-item ${isExpired ? 'expired' : ''}`}>
      <div className="url-list-item-header">
        <a
          href={shortUrl}
          target="_blank"
          rel="noopener noreferrer"
          className="url-list-short-link"
        >
          {shortUrl}
        </a>
        <div className="url-list-actions">
          <span className={`url-list-time ${isExpired ? 'expired' : ''}`}>
            {isExpired ? 'expired' : timeRemaining}
          </span>
          <button
            onClick={() => onCopy(url.short_code, shortUrl)}
            className={`url-list-btn url-list-btn-copy ${isCopied ? 'copied' : ''}`}
            title="Copy"
          >
            {isCopied ? 'Copied' : 'Copy'}
          </button>
          <button
            onClick={handleDelete}
            className="url-list-btn url-list-btn-delete"
            title="Delete"
          >
            Delete
          </button>
        </div>
      </div>
      <a
        href={url.original_url}
        target="_blank"
        rel="noopener noreferrer"
        className="url-list-original"
      >
        {truncateUrl(url.original_url)}
      </a>
    </div>
  );
}

export function UrlList({ refreshTrigger }) {
  const [urls, setUrls] = useState([]);
  const [loading, setLoading] = useState(true);
  const [copiedCode, setCopiedCode] = useState(null);

  const loadUrls = async () => {
    try {
      const data = await getUserUrls();
      setUrls(data.urls || []);
    } catch (error) {
      console.error('Failed to load URLs:', error);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadUrls();
  }, [refreshTrigger]);

  const handleDelete = (shortCode) => {
    setUrls(urls.filter(u => u.short_code !== shortCode));
  };

  const handleCopy = (shortCode, shortUrl) => {
    navigator.clipboard.writeText(shortUrl);
    setCopiedCode(shortCode);
    setTimeout(() => setCopiedCode(null), 2000);
  };

  if (loading) {
    return null;
  }

  if (urls.length === 0) {
    return null;
  }

  return (
    <div className="url-list-section">
      <h2 className="url-list-title">Your URLs</h2>
      <div className="url-list">
        {urls.map((url) => (
          <UrlItem
            key={url.short_code}
            url={url}
            onDelete={handleDelete}
            onCopy={handleCopy}
            copiedCode={copiedCode}
          />
        ))}
      </div>
    </div>
  );
}
