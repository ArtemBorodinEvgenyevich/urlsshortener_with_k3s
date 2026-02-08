const API_BASE_URL = import.meta.env.VITE_API_URL || '';
const URL_TTL = 43200;

export const initSession = async () => {
  try {
    const response = await fetch(`${API_BASE_URL}/auth/session`, {
      method: 'POST',
      credentials: 'include',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ metadata: {} }),
    });

    if (!response.ok) {
      console.error('Failed to init session:', response.status);
      return null;
    }

    return await response.json();
  } catch (error) {
    console.error('Session init error:', error);
    return null;
  }
};

export const shortenUrl = async (url) => {
  const response = await fetch(`${API_BASE_URL}/api/v1/shorten`, {
    method: 'POST',
    credentials: 'include',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ url, ttl: URL_TTL }),
  });

  if (!response.ok) {
    throw new Error(`HTTP error! status: ${response.status}, ${response.statusText}`);
  }

  return await response.json();
};

export const getUrl = async (shortCode) => {
  const response = await fetch(`${API_BASE_URL}/api/v1/urls/${shortCode}`, {
    method: 'GET',
    credentials: 'include',
  });

  if (!response.ok) {
    throw new Error(`HTTP error! status: ${response.status}, ${response.statusText}`);
  }

  return await response.json();
};

export const getUserUrls = async (limit = 20, offset = 0) => {
  const response = await fetch(`${API_BASE_URL}/api/v1/urls?limit=${limit}&offset=${offset}`, {
    method: 'GET',
    credentials: 'include',
  });

  if (!response.ok) {
    if (response.status === 401) {
      return { urls: [] };
    }
    throw new Error(`HTTP error! status: ${response.status}, ${response.statusText}`);
  }

  return await response.json();
};

export const deleteUrl = async (shortCode) => {
  const response = await fetch(`${API_BASE_URL}/api/v1/urls/${shortCode}`, {
    method: 'DELETE',
    credentials: 'include',
  });

  if (!response.ok) {
    throw new Error(`HTTP error! status: ${response.status}, ${response.statusText}`);
  }

  return true;
};

export default {
  initSession,
  shortenUrl,
  getUrl,
  getUserUrls,
  deleteUrl,
};
