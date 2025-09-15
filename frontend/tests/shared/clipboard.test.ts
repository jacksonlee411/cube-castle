// @vitest-environment jsdom
import { describe, it, expect, vi } from 'vitest';
import { copyText } from '../../src/shared/utils/clipboard';

describe('copyText', () => {
  it('uses navigator.clipboard when available', async () => {
    const writeText = vi.fn().mockResolvedValue(undefined);
    Object.assign(navigator, { clipboard: { writeText } });
    const ok = await copyText('hello');
    expect(ok).toBe(true);
    expect(writeText).toHaveBeenCalledWith('hello');
  });

  it('falls back to execCommand when clipboard API not available', async () => {
    // Remove clipboard and mock execCommand
    // @ts-expect-error override for test
    delete navigator.clipboard;
    (document as any).execCommand = vi.fn().mockReturnValue(true);
    const ok = await copyText('world');
    expect(ok).toBe(true);
    expect((document as any).execCommand).toHaveBeenCalledWith('copy');
  });
});
