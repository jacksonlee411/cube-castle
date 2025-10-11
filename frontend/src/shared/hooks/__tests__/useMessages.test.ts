import { describe, expect, it, vi, afterEach } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { useMessages } from '../useMessages';

describe('useMessages', () => {
  afterEach(() => {
    vi.useRealTimers();
  });

  it('shows success message and clears automatically', () => {
    vi.useFakeTimers();
    const { result } = renderHook(() => useMessages());

    act(() => {
      result.current.showSuccess('保存成功');
    });

    expect(result.current.successMessage).toBe('保存成功');
    expect(result.current.error).toBeNull();

    act(() => {
      vi.advanceTimersByTime(3000);
    });

    expect(result.current.successMessage).toBeNull();
  });

  it('shows error message and clears automatically', () => {
    vi.useFakeTimers();
    const { result } = renderHook(() => useMessages());

    act(() => {
      result.current.showError('网络异常');
    });

    expect(result.current.error).toBe('网络异常');
    expect(result.current.successMessage).toBeNull();

    act(() => {
      vi.advanceTimersByTime(5000);
    });

    expect(result.current.error).toBeNull();
  });

  it('clears messages immediately', () => {
    vi.useFakeTimers();
    const { result } = renderHook(() => useMessages());

    act(() => {
      result.current.showSuccess('已提交');
      result.current.showError('提交失败');
    });

    expect(result.current.error).toBe('提交失败');

    act(() => {
      result.current.clearMessages();
    });

    expect(result.current.error).toBeNull();
    expect(result.current.successMessage).toBeNull();
  });
});
