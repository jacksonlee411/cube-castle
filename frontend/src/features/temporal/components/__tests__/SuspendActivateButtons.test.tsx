import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { vi } from 'vitest';
import { SuspendActivateButtons } from '../SuspendActivateButtons';

const suspendSpy = vi.fn();
const activateSpy = vi.fn();

vi.mock('@workday/canvas-kit-react/layout', () => ({
  Flex: ({ children }: { children: React.ReactNode }) => <div>{children}</div>
}));

vi.mock('@workday/canvas-kit-react/button', () => ({
  SecondaryButton: ({ children, ...props }: any) => <button {...props}>{children}</button>,
  PrimaryButton: ({ children, ...props }: any) => <button {...props}>{children}</button>
}));

vi.mock('@workday/canvas-kit-react/icon', () => ({
  SystemIcon: ({ children }: any) => <span>{children}</span>
}));

vi.mock('@workday/canvas-kit-react/modal', () => {
  const ModalComponent: any = ({ children }: { children: React.ReactNode }) => <div>{children}</div>;
  ModalComponent.Overlay = ({ children }: { children: React.ReactNode }) => <div>{children}</div>;
  ModalComponent.Card = ({ children }: { children: React.ReactNode }) => <div>{children}</div>;
  ModalComponent.CloseIcon = ({ onClick }: { onClick?: () => void }) => (
    <button onClick={onClick} aria-label="关闭">
      ×
    </button>
  );
  ModalComponent.Heading = ({ children }: { children: React.ReactNode }) => <h2>{children}</h2>;
  ModalComponent.Body = ({ children }: { children: React.ReactNode }) => <div>{children}</div>;
  return {
    Modal: ModalComponent,
    useModalModel: () => ({ events: { show: vi.fn(), hide: vi.fn() } })
  };
});

vi.mock('@workday/canvas-kit-react/text', () => ({
  Text: ({ children }: { children: React.ReactNode }) => <p>{children}</p>
}));

vi.mock('@workday/canvas-kit-react/text-input', () => ({
  TextInput: ({ children, ...props }: any) => <input {...props}>{children}</input>
}));

vi.mock('@workday/canvas-kit-react/text-area', () => ({
  TextArea: ({ children, ...props }: any) => <textarea {...props}>{children}</textarea>
}));

vi.mock('@workday/canvas-kit-react/common', () => ({ CanvasProvider: ({ children }: any) => <>{children}</> }));

vi.mock('@workday/canvas-kit-react/tokens', () => ({ colors: {} }));

vi.mock('@workday/canvas-system-icons-web', () => ({
  mediaPauseIcon: 'pause-icon',
  mediaPlayIcon: 'play-icon'
}));

vi.mock('@/shared/hooks/useOrganizationMutations', () => ({
  useSuspendOrganization: () => ({ mutateAsync: suspendSpy, isPending: false }),
  useActivateOrganization: () => ({ mutateAsync: activateSpy, isPending: false })
}));

const renderComponent = (props: Partial<React.ComponentProps<typeof SuspendActivateButtons>> = {}) => {
  const defaultProps = {
    organizationCode: '1000004',
    currentStatus: 'ACTIVE' as const,
    currentETag: '123',
    readonly: false,
    disabled: false,
    onETagChange: vi.fn(),
    onSuccess: vi.fn(),
    onError: vi.fn(),
    onCompleted: vi.fn()
  };
  return render(<SuspendActivateButtons {...defaultProps} {...props} />);
};

describe('SuspendActivateButtons', () => {
  beforeEach(() => {
    suspendSpy.mockReset();
    activateSpy.mockReset();
  });

  it('allows selecting date and reason before suspending', async () => {
    const onETagChange = vi.fn();
    const onCompleted = vi.fn();
    const onSuccess = vi.fn();
    const onError = vi.fn();

    suspendSpy.mockResolvedValue({ organization: {}, etag: 'etag', headers: {} });

    renderComponent({ onETagChange, onCompleted, onSuccess, onError });

    fireEvent.click(screen.getByText('暂停组织'));

    await waitFor(() => expect(screen.getByTestId('status-change-date-input')).toBeInTheDocument());
    const dateInput = screen.getByTestId('status-change-date-input') as HTMLInputElement;

    fireEvent.change(dateInput, { target: { value: '2025-10-15' } });
    fireEvent.change(screen.getByTestId('status-change-reason-input'), {
      target: { value: '年度盘点' }
    });

    fireEvent.click(screen.getByTestId('status-change-confirm'));

    await waitFor(() => expect(suspendSpy).toHaveBeenCalledTimes(1));

    expect(suspendSpy).toHaveBeenCalledWith({
      code: '1000004',
      effectiveDate: '2025-10-15',
      currentETag: '123',
      operationReason: '年度盘点'
    });
    expect(onETagChange).toHaveBeenCalledWith('etag');
    expect(onCompleted).toHaveBeenCalled();
    expect(onSuccess).toHaveBeenCalledWith('组织已停用');
    expect(onError).not.toHaveBeenCalled();
  });

  it('triggers activate flow when current status is INACTIVE', async () => {
    const onCompleted = vi.fn();
    activateSpy.mockResolvedValue({ organization: {}, etag: 'etag2', headers: {} });

    renderComponent({ currentStatus: 'INACTIVE', currentETag: null, onCompleted });

    fireEvent.click(screen.getByText('重新启用'));
    fireEvent.click(await screen.findByTestId('status-change-confirm'));

    await waitFor(() => expect(activateSpy).toHaveBeenCalledTimes(1));
    expect(onCompleted).toHaveBeenCalledWith('activate', { organization: {}, etag: 'etag2', headers: {} });
    expect(suspendSpy).not.toHaveBeenCalled();
  });
});
