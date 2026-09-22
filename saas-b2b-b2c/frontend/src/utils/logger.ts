// utils/logger.ts — централизованный логгер для прода (Sentry-ready)
type LogLevel = 'debug' | 'info' | 'warn' | 'error';
const isProd = process.env.NODE_ENV === 'production';

function log(level: LogLevel, ...args: unknown[]) {
  if (isProd && level === 'debug') return;
  // в проде error/warn идут в консоль, в будущем заменить на Sentry.captureMessage
  if (level === 'error') console.error(...args);
  else if (level === 'warn') console.warn(...args);
  else if (level === 'debug' && !isProd) console.debug(...args);
  else if (level === 'info' && !isProd) console.info(...args);
}

export const logger = {
  debug: (...args: unknown[]) => log('debug', ...args),
  info: (...args: unknown[]) => log('info', ...args),
  warn: (...args: unknown[]) => log('warn', ...args),
  error: (...args: unknown[]) => log('error', ...args),
};
export default logger;
