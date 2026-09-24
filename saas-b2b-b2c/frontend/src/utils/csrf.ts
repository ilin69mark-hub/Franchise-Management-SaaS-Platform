// F7: CSRF double-submit — читаем csrf_token cookie (её ставит бэкенд
// при login/register/refresh, см. middleware.CSRF) и шлём заголовком
// X-CSRF-Token на мутации. Сессионные cookie браузер отправляет сам
// (withCredentials / credentials:include), токены в localStorage НЕ храним.

export const CSRF_COOKIE = 'csrf_token';
export const CSRF_HEADER = 'X-CSRF-Token';

export function getCsrfToken(): string | null {
  if (typeof document === 'undefined') return null;
  const parts = document.cookie ? document.cookie.split(';') : [];
  for (const part of parts) {
    const idx = part.indexOf('=');
    if (idx < 0) continue;
    const name = part.slice(0, idx).trim();
    if (name === CSRF_COOKIE) {
      const value = part.slice(idx + 1).trim();
      try {
        return decodeURIComponent(value);
      } catch {
        return value;
      }
    }
  }
  return null;
}

/** Заголовки для мутаций через axios-like конфиги. */
export function csrfHeader(): Record<string, string> {
  const token = getCsrfToken();
  return token ? { [CSRF_HEADER]: token } : {};
}
