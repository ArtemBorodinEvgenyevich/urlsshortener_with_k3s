import { useState, useCallback } from 'react'

export const useCopyToClipboard = (resetDelay = 2000) => {
  const [isCopied, setIsCopied] = useState(false);

  const copyToClipboard = useCallback(async (text) => {
    if (!text) {
      console.warn("Nothing to copy");
      return false;
    }

    try {
      if (navigator.clipboard && window.isSecureContext) {
        await navigator.clipboard.writeText(text);
      } else {
        const textArea = document.createElement('textarea');
        textArea.value = text;
        textArea.style.position = 'fixed';
        textArea.style.left = '-999999px';
        textArea.style.top = '-999999px';
        document.body.appendChild(textArea);
        textArea.focus();
        textArea.select();

        try {
          document.execCommand('copy');
          textArea.remove();
        } catch (err) {
          console.error('Fallback: Oops, unable to copy', err);
          textArea.remove();
          return false;
        }
      }

      setIsCopied(true);
      setTimeout(() => setIsCopied(false), resetDelay);
      return true;
    } catch (err) {
      console.error('Failed to copy:', err);
      setIsCopied(false);
      return false;
    }
  }, [resetDelay]);

  return { isCopied, copyToClipboard };
}
