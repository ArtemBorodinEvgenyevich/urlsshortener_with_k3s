import { useEffect, useState } from 'react';
import { useParams } from 'react-router-dom';
import { getUrl } from '../services/api.js';

export function RedirectPage() {
  const { shortCode } = useParams();
  const [error, setError] = useState(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchAndRedirect = async () => {
      try {
        const data = await getUrl(shortCode);
        window.location.replace(data.original_url);
      } catch (err) {
        console.error('Error:', err);
        if (err.message.includes('404')) {
          setError('Короткая ссылка не найдена');
        } else {
          setError('Произошла ошибка');
        }
        setLoading(false);
      }
    };

    if (shortCode) {
      fetchAndRedirect();
    }
  }, [shortCode]);

  if (loading) {
    return (
      <div className="app">
        <div className="pure-g">
          <div className="pure-u-1 pure-u-md-22-24 pure-u-lg-18-24 center-container">
            <header className="header">
              <h1 className="title">
                <span className="gradient-text">URL Shortener</span>
              </h1>
            </header>

            <main className="main" style={{ textAlign: 'center' }}>
              <div className="result-card">
                <div style={{ marginBottom: '2rem' }}>
                  <div className="spinner" style={{
                    width: '60px',
                    height: '60px',
                    border: '4px solid #e0e7ff',
                    borderTopColor: '#667eea',
                    borderRadius: '50%',
                    animation: 'spin 0.8s linear infinite',
                    margin: '0 auto'
                  }}></div>
                </div>
                <h2 className="result-title">Перенаправление...</h2>
                <p style={{ color: '#64748b', fontSize: '1.125rem' }}>
                  Подождите, мы перенаправляем вас на оригинальный URL
                </p>
              </div>
            </main>

            <footer className="footer">
              <p>Создано с ❤️ для сокращения ссылок</p>
            </footer>
          </div>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="app">
        <div className="pure-g">
          <div className="pure-u-1 pure-u-md-22-24 pure-u-lg-18-24 center-container">
            <header className="header">
              <h1 className="title">
                <span className="gradient-text">URL Shortener</span>
              </h1>
            </header>

            <main className="main" style={{ textAlign: 'center' }}>
              <div className="alert alert-error">
                <div style={{ fontSize: '3rem', marginBottom: '1rem' }}>⚠️</div>
                <h2 style={{ fontSize: '1.5rem', fontWeight: '700', margin: '0 0 0.5rem 0' }}>
                  Ошибка
                </h2>
                <p style={{ margin: '0 0 1.5rem 0' }}>{error}</p>
                <a
                  href="/"
                  className="pure-button copy-button"
                  style={{
                    display: 'inline-flex',
                    width: 'auto',
                    minWidth: '200px'
                  }}
                >
                  <span>🏠</span>
                  <span>На главную</span>
                </a>
              </div>
            </main>

            <footer className="footer">
              <p>Создано с ❤️ для сокращения ссылок</p>
            </footer>
          </div>
        </div>
      </div>
    );
  }

  return null;
}
