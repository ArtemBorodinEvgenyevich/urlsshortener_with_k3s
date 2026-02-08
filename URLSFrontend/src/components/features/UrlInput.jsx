import { useUrlInput } from "../../hooks/useUrlInput.js";
import './UrlInput.css'

export const UrlInput = ({ onSubmit }) => {
  const {
    url,
    error,
    isLoading,
    isValid,
    touched,
    handleChange,
    handleBlur,
    validate,
    getNormalizedUrl,
    setIsLoading,
  } = useUrlInput();

  const handleSubmit = async (e) => {
    e.preventDefault();

    if (!validate()) {
      return;
    }

    const normalizedUrl = getNormalizedUrl();

    try {
      setIsLoading(true);
      await onSubmit(normalizedUrl);
    } catch (err) {
      console.error('Error shortening URL:', err);
    } finally {
      setIsLoading(false);
    }
  };

  const showError = touched && error;

  return (
      <div className="url-input-container">
        {/* Pure CSS Form */}
        <form onSubmit={handleSubmit} className="pure-form url-input-form" noValidate>

          {/* Pure CSS Grid for form */}
          <div className="pure-g">
            <div className="pure-u-1 pure-u-md-17-24">
              <div className="input-wrapper">
                <input
                    type="url"
                    value={url}
                    onChange={handleChange}
                    onBlur={handleBlur}
                    placeholder="Input your URL..."
                    className={`pure-input-1 url-input ${showError ? 'error' : ''} ${isValid && url ? 'valid' : ''}`}
                    disabled={isLoading}
                    aria-label="URL for shortening"
                    aria-invalid={showError ? 'true' : 'false'}
                    aria-describedby={showError ? 'url-error' : undefined}
                    autoFocus
                />

                {showError && (
                    <span id="url-error" className="error-message" role="alert">
                    {error}
                  </span>
                )}
              </div>
            </div>

            <div className="pure-u-1 pure-u-md-7-24">
              <button
                  type="submit"
                  className="pure-button pure-button-primary submit-button"
                  disabled={!isValid || isLoading}
                  aria-label="Shorten URL"
              >
                {isLoading ? (
                    <>
                      <span className="spinner" aria-hidden="true"></span>
                      <span>Сокращаем...</span>
                    </>
                ) : (
                    'Create short url'
                )}
              </button>
            </div>
          </div>
        </form>
      </div>
  );
};
