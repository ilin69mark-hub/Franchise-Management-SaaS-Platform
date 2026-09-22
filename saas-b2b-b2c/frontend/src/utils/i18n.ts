// utils/i18n.ts — заглушка i18n для v1.3 (ru-only, без next-intl зависимости)
// Полноценный next-intl — 40-60ч (1461 строк). Сейчас — структура + 10 ключей ru.json, t() возвращает key fallback.
// След шаг: npm i next-intl + provider в _app.tsx + замена хардкода 1461 → t('common.save')
import ru from '@/locales/ru.json';

type NestedKeyOf<T> = T extends object ? { [K in keyof T]: K extends string ? `${K}` | `${K}.${NestedKeyOf<T[K]>}` : never }[keyof T] : never;
const dict: Record<string, unknown> = ru as Record<string, unknown>;

function get(obj: Record<string, unknown>, path: string): string | undefined {
  const parts = path.split('.');
  let cur: unknown = obj;
  for (const p of parts) {
    if (cur && typeof cur === 'object' && p in (cur as Record<string, unknown>)) cur = (cur as Record<string, unknown>)[p];
    else return undefined;
  }
  return typeof cur === 'string' ? cur : undefined;
}

export function t(key: string, fallback?: string): string {
  return get(dict, key) ?? fallback ?? key;
}
export default t;
