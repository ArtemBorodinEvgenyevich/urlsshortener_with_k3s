
export const isValidUrl = (url) => {
  if (!url || typeof url !== 'string') {
    return false;
  }

  const trimmedUrl = url.trim();
  if (trimmedUrl.length === 0) {
    return false;
  }

  try {
    const urlObject = new URL(trimmedUrl);
    return urlObject.protocol === 'https:' || urlObject.protocol === 'http:';
  } catch (error) {
    try {
      const urlWithProtoc = new URL(`https://${trimmedUrl}`);
      return urlWithProtoc.hostname.includes('.');
    } catch {
      return false;
    }
  }
};

export const normalizeURL = (url) => {
  if (!url || typeof url !== 'string') {
    return '';
  }

  const trimmedUrl = url.trim();
  if (!trimmedUrl) {
    return '';
  }

  if (/^https?:\/\//i.test(trimmedUrl)) {
    return trimmedUrl;
  }

  return `https://${trimmedUrl}`;
}

export const getUrlError = (url) => {
  if (!url || url.trim().length === 0) {
    return 'URL cannot be empty';
  }

  if (!isValidUrl(url)) {
    return "Provide a valid URL (Ex. https://example.com)";
  }

  const normalizedUrl = normalizeURL(url);

  if (normalizedUrl.length > 2048) {
    return "URL is too long (maximum of 2048 characters)"
  }

  return null;
}
