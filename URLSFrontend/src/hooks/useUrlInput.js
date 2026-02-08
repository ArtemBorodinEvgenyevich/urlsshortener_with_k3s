import { useState, useCallback } from 'react';
import { isValidUrl, normalizeURL, getUrlError } from "../utils/validation.js";

export const useUrlInput = () => {
  const [url, setUrl] = useState("");
  const [error, setError] = useState(null);
  const [isLoading, setIsLoading] = useState(false);
  const [touched, setTouched] = useState(false);

  // input handler
  const handleChange = useCallback((e) => {
    const value = e.target.value;
    setUrl(value);

    if (error) {
      setError(null);
    }
  }, [error]);

  const handleBlur = useCallback(() => {
    setTouched(true);

    if (url.trim()) {
      const validationError = getUrlError();
      setError(validationError);
    }
  }, [url]);

  const validate = useCallback(() => {
    const validationError = getUrlError(url);
    setError(validationError);
    return !validationError;
  }, [url]);

  const getNormalizedUrl = useCallback(() => {
    return normalizeURL(url);
  }, [url]);

  const reset = useCallback(() => {
    setUrl("");
    setError(null);
    setIsLoading(false);
    setTouched(false);
  }, []);

  const isValid = isValidUrl(url);

  return {
    // State
    url,
    error,
    isLoading,
    isValid,
    touched,

    // Handlers
    handleChange,
    handleBlur,
    setIsLoading,

    // Methods
    validate,
    getNormalizedUrl,
    reset,
  };
};
