import viteLogo from '/vite.svg'
import { useState, useEffect } from 'react'
import { Routes, Route } from 'react-router-dom'
import { useCopyToClipboard } from "./hooks/useCopyToClipboard.js";
import { shortenUrl, initSession } from "./services/api.js";
import { UrlInput } from "./components/features/UrlInput.jsx";
import { UrlList } from "./components/features/UrlList.jsx";
import { RedirectPage } from "./pages/RedirectPage.jsx";
import './App.css'

function HomePage() {
  const [shortenedUrl, setShortenedUrl] = useState(null);
  const [error, setError] = useState(null);
  const [refreshTrigger, setRefreshTrigger] = useState(0);
  const { isCopied, copyToClipboard } = useCopyToClipboard(2000);

  const handleUrlSubmit = async (url) => {
    try {
      setError(null);
      const response = await shortenUrl(url);

      // Формируем полный URL с текущим хостом
      const shortUrl = `${window.location.origin}/${response.short_url}`;

      setShortenedUrl({
        originalUrl: url,
        shortUrl: shortUrl,
        shortCode: response.short_url,
      });
      setRefreshTrigger(prev => prev + 1);
    } catch (error) {
      setError("Ошибка при сокращении URL");
      console.error("Error:", error);
    }
  };

  const handleCopy = () => {
    if (shortenedUrl?.shortUrl) {
      copyToClipboard(shortenedUrl.shortUrl);
    }
  };

  return (
      <div className="app">
        <div className="pure-g">
          <div className="pure-u-1 pure-u-md-22-24 pure-u-lg-18-24 center-container">
            <header className="header">
              <h1 className="title">
                <span className="gradient-text">URL Shortener</span>
              </h1>
              <p className="subtitle">Simple url shortener</p>
            </header>

            <main className="main">
              <UrlInput onSubmit={handleUrlSubmit} />

              {error && (
                  <div className="alert alert-error" role="alert">
                    {error}
                  </div>
              )}

              {shortenedUrl && (
                  <div className="result-card">
                    <h2 className="result-title">Your short url is: </h2>
                    <div className="pure-g url-display-grid">
                      <div className="pure-u-1 pure-u-md-16-24">
                        <a
                            href={shortenedUrl.shortUrl}
                            target="_blank"
                            rel="noopener noreferrer"
                            className="short-url"
                        >
                          {shortenedUrl.shortUrl}
                        </a>
                      </div>
                      <div className="pure-u-1 pure-u-md-8-24">
                        <button
                            onClick={handleCopy}
                            className={`pure-button copy-button ${isCopied ? 'copied' : ''}`}
                        >
                          {isCopied ? (
                              <>
                                <span className="checkmark">✓</span>
                                <span>Copied!</span>
                              </>
                          ) : (
                              <>
                                <span>📋</span>
                                <span>Copy</span>
                              </>
                          )}
                        </button>
                      </div>
                    </div>
                    <div className="original-url">
                      <span className="label">Original url:</span>
                      <span className="url-text">{shortenedUrl.originalUrl}</span>
                    </div>
                  </div>
              )}

              <UrlList refreshTrigger={refreshTrigger} />
            </main>

            <footer className="footer">
              <p>Создано с ❤️ для сокращения ссылок</p>
            </footer>
          </div>
        </div>
      </div>
  );
}

function App() {
  useEffect(() => {
    initSession();
  }, []);

  return (
      <Routes>
        <Route path="/" element={<HomePage />} />
        <Route path="/:shortCode" element={<RedirectPage />} />
      </Routes>
  );
}

export default App

