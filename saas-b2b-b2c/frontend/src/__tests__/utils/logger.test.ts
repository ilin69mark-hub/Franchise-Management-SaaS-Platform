import logger from '@/utils/logger';

describe('logger', () => {
  const origError = console.error;
  const origWarn = console.warn;
  const origDebug = console.debug;
  const origInfo = console.info;
  beforeEach(() => {
    jest.spyOn(console, 'error').mockImplementation(() => {});
    jest.spyOn(console, 'warn').mockImplementation(() => {});
    jest.spyOn(console, 'debug').mockImplementation(() => {});
    jest.spyOn(console, 'info').mockImplementation(() => {});
  });
  afterEach(() => {
    (console.error as jest.Mock).mockRestore();
    (console.warn as jest.Mock).mockRestore();
    (console.debug as jest.Mock).mockRestore();
    (console.info as jest.Mock).mockRestore();
  });

  it('error calls console.error', () => {
    logger.error('a', 1);
    expect(console.error).toHaveBeenCalledWith('a', 1);
  });
  it('warn calls console.warn', () => {
    logger.warn('w');
    expect(console.warn).toHaveBeenCalledWith('w');
  });
  it('debug calls console.debug in dev', () => {
    (process.env.NODE_ENV as string) = 'development';
    logger.debug('d');
    expect(console.debug).toHaveBeenCalledWith('d');
  });
  it('info calls console.info', () => {
    (process.env.NODE_ENV as string) = 'development';
    logger.info('i');
    expect(console.info).toHaveBeenCalledWith('i');
  });
});
