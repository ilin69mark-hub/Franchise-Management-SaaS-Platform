import { t } from '@/utils/i18n';

describe('i18n t', () => {
  it('returns ru translation for common.save', () => {
    expect(t('common.save')).toBe('Сохранить');
  });
  it('returns fallback if key missing', () => {
    expect(t('missing.key', 'fb')).toBe('fb');
  });
  it('returns key if missing and no fallback', () => {
    expect(t('missing.key2')).toBe('missing.key2');
  });
  it('handles nested dashboard key', () => {
    expect(t('dashboard.territoryCommunications')).toBe('Коммуникации территории');
  });
});
